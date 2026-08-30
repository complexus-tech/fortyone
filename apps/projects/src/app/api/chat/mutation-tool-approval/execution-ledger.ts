import "server-only";

import {
  claimMutationApprovalExecution,
  completeMutationApprovalExecution,
  failMutationApprovalExecution,
  startMutationApprovalExecution,
} from "@/modules/ai-chats/actions/mutation-approval-execution";
import type {
  MutationApprovalExecution,
  MutationApprovalFailureCode,
} from "@/modules/ai-chats/actions/mutation-approval-execution";
import { getChatErrorDiagnostic } from "../chat-errors";
import {
  getApprovalFingerprint,
  getPreparedApprovalFingerprint,
  type MutationToolApproval,
} from "./approval-fingerprint";
import {
  prepareApprovedMutation,
  type PreparedApprovedMutation,
} from "./approval-policy";
import {
  createExecutionDeadlineSignal,
  executeApprovedMutation,
  getApprovalAbortError,
  isFailedMutationOutput,
  throwIfRequestAborted,
  waitForAbort,
} from "./tool-execution";

const PERSISTED_APPROVAL_RETRY_DELAYS_MS = [
  0, 50, 100, 200, 400, 800, 1600,
] as const;
const APPROVAL_EXECUTION_TIMEOUT_MS = 110_000;
const APPROVAL_LEDGER_WAIT_TIMEOUT_MS = 115_000;
const APPROVAL_POLL_DELAYS_MS = [
  100, 250, 500, 1000, 2000, 3000, 5000,
] as const;
const READY_LEASE_RECLAIM_GRACE_MS = 25;
const MAX_START_ATTEMPTS = 3;
export type ApprovalExecutionResult =
  | { denied: true }
  | {
      denied: false;
      durableFingerprint?: string;
      haltsFollowing: boolean;
      output: unknown;
    };

type PendingApprovalExecution = {
  fingerprint: string;
  result: Promise<ApprovalExecutionResult>;
};

const pendingApprovalExecutions = new Map<string, PendingApprovalExecution>();

const getApprovalFailureOutput = (error: string) => ({
  error,
  success: false,
});

const getErrorStatus = (error: unknown) => {
  if (!error || typeof error !== "object" || !("status" in error)) {
    return undefined;
  }

  return typeof error.status === "number" ? error.status : undefined;
};

const waitForApprovalPersistence = (
  delayMs: number,
  abortSignal: AbortSignal,
) =>
  new Promise<void>((resolve, reject) => {
    if (abortSignal.aborted) {
      reject(getApprovalAbortError(abortSignal));
      return;
    }

    const timeout = setTimeout(() => {
      abortSignal.removeEventListener("abort", onAbort);
      resolve();
    }, delayMs);
    const onAbort = () => {
      clearTimeout(timeout);
      reject(getApprovalAbortError(abortSignal));
    };
    abortSignal.addEventListener("abort", onAbort, { once: true });
  });

const claimMutationApprovalEventually = async ({
  abortSignal,
  approval,
  chatId,
  fingerprint,
  workspaceSlug,
}: {
  abortSignal: AbortSignal;
  approval: MutationToolApproval;
  chatId: string;
  fingerprint: string;
  workspaceSlug: string;
}) => {
  let lastNotFoundError: unknown;

  for (const delayMs of PERSISTED_APPROVAL_RETRY_DELAYS_MS) {
    if (delayMs > 0) {
      // eslint-disable-next-line no-await-in-loop -- The first approval can race initial chat persistence.
      await waitForApprovalPersistence(delayMs, abortSignal);
    }

    try {
      // eslint-disable-next-line no-await-in-loop -- A durable claim must exist before the mutation can execute.
      return await claimMutationApprovalExecution({
        chatId,
        fingerprint,
        toolCallId: approval.toolCallId,
        workspaceSlug,
      });
    } catch (error) {
      if (getErrorStatus(error) !== 404) throw error;
      lastNotFoundError = error;
    }
  }

  if (lastNotFoundError instanceof Error) throw lastNotFoundError;
  throw new Error("The prepared chat action was not found.");
};

const isPendingMutationApproval = (execution: MutationApprovalExecution) =>
  execution.state === "executing" || execution.state === "ready";

const getPendingPollDelay = (
  execution: MutationApprovalExecution,
  pollIndex: number,
) => {
  const backoff =
    APPROVAL_POLL_DELAYS_MS[
      Math.min(pollIndex, APPROVAL_POLL_DELAYS_MS.length - 1)
    ];
  if (execution.state !== "ready" || !execution.leaseExpiresAt) {
    return backoff;
  }

  const leaseExpiresAt = Date.parse(execution.leaseExpiresAt);
  if (!Number.isFinite(leaseExpiresAt)) return backoff;
  const leaseRemainingMs =
    leaseExpiresAt - Date.now() + READY_LEASE_RECLAIM_GRACE_MS;
  return Math.max(
    READY_LEASE_RECLAIM_GRACE_MS,
    Math.min(backoff, leaseRemainingMs),
  );
};

const waitForActionableMutationApproval = async ({
  abortSignal,
  approval,
  chatId,
  deadlineAt,
  fingerprint,
  workspaceSlug,
}: {
  abortSignal: AbortSignal;
  approval: MutationToolApproval;
  chatId: string;
  deadlineAt: number;
  fingerprint: string;
  workspaceSlug: string;
}): Promise<MutationApprovalExecution> => {
  let pollIndex = 0;
  let execution = await claimMutationApprovalEventually({
    abortSignal,
    approval,
    chatId,
    fingerprint,
    workspaceSlug,
  });

  while (isPendingMutationApproval(execution) && Date.now() < deadlineAt) {
    const remainingMs = deadlineAt - Date.now();

    const delayMs = Math.min(
      remainingMs,
      getPendingPollDelay(execution, pollIndex),
    );
    pollIndex += 1;
    // eslint-disable-next-line no-await-in-loop -- Polling is bounded by deadlineAt and observes one authoritative ledger state at a time.
    await waitForApprovalPersistence(delayMs, abortSignal);
    // eslint-disable-next-line no-await-in-loop -- Durable state must be observed serially before execution can be reclaimed.
    execution = await claimMutationApprovalEventually({
      abortSignal,
      approval,
      chatId,
      fingerprint,
      workspaceSlug,
    });
  }

  return execution;
};

const getUncertainApprovalResult = (): ApprovalExecutionResult => ({
  denied: false,
  haltsFollowing: true,
  output: getApprovalFailureOutput(
    "Maya could not verify whether this approved change finished. Check the workspace before trying it again; an identical change is blocked until this execution is reconciled.",
  ),
});

const getPendingApprovalResult = (): ApprovalExecutionResult => ({
  denied: false,
  haltsFollowing: true,
  output: getApprovalFailureOutput(
    "This approved change is still being processed. Maya did not run it again. Wait for the current operation to finish before preparing another identical change.",
  ),
});

const getCompletedApprovalResult = (
  fingerprint: string,
  output: unknown,
): ApprovalExecutionResult => ({
  denied: false,
  durableFingerprint: fingerprint,
  haltsFollowing: isFailedMutationOutput(output),
  output,
});

const failMutationApprovalSafely = async ({
  approval,
  chatId,
  failureCode,
  fingerprint,
  leaseToken,
  workspaceSlug,
}: {
  approval: MutationToolApproval;
  chatId: string;
  failureCode: MutationApprovalFailureCode;
  fingerprint: string;
  leaseToken: string;
  workspaceSlug: string;
}): Promise<MutationApprovalExecution | undefined> => {
  try {
    return await failMutationApprovalExecution({
      chatId,
      failureCode,
      fingerprint,
      leaseToken,
      toolCallId: approval.toolCallId,
      workspaceSlug,
    });
  } catch (error) {
    // eslint-disable-next-line no-console -- Do not log the approved payload while preserving recovery diagnostics.
    console.error(
      `[chat/route] Could not quarantine ${approval.toolName} approval`,
      getChatErrorDiagnostic(error),
    );
    // Either the failure transition or a competing completion may have
    // committed despite the lost response. Re-read the same ledger identity
    // before emitting an uncertainty receipt so completed output remains
    // authoritative and a later tool-call ID cannot replay the mutation.
    try {
      return await claimMutationApprovalExecution({
        chatId,
        fingerprint,
        toolCallId: approval.toolCallId,
        workspaceSlug,
      });
    } catch (readError) {
      // eslint-disable-next-line no-console -- Do not log the approved payload while preserving recovery diagnostics.
      console.error(
        `[chat/route] Could not re-read ${approval.toolName} approval after a lost failure response`,
        getChatErrorDiagnostic(readError),
      );
      return undefined;
    }
  }
};

const executeApprovalWithLedger = async ({
  abortSignal,
  approval,
  chatId,
  workspaceSlug,
}: {
  abortSignal: AbortSignal;
  approval: MutationToolApproval;
  chatId: string;
  workspaceSlug: string;
}): Promise<ApprovalExecutionResult> => {
  if (!approval.approved) return { denied: true };
  throwIfRequestAborted(abortSignal);

  let prepared: PreparedApprovedMutation;
  try {
    prepared = await prepareApprovedMutation(approval);
  } catch (error) {
    // eslint-disable-next-line no-console -- The payload is intentionally omitted from validation diagnostics.
    console.error(
      `[chat/route] Could not validate ${approval.toolName} approval`,
      getChatErrorDiagnostic(error),
    );
    return {
      denied: false,
      haltsFollowing: true,
      output: getApprovalFailureOutput(
        "The approved change could not be validated and was not executed. Ask Maya to prepare it again.",
      ),
    };
  }

  const fingerprint = getPreparedApprovalFingerprint(approval, prepared.input);

  try {
    throwIfRequestAborted(abortSignal);
    const ledgerDeadlineAt = Date.now() + APPROVAL_LEDGER_WAIT_TIMEOUT_MS;
    let claim = await waitForActionableMutationApproval({
      abortSignal,
      approval,
      chatId,
      deadlineAt: ledgerDeadlineAt,
      fingerprint,
      workspaceSlug,
    });
    let leaseToken: string | undefined;

    for (
      let startAttempt = 0;
      startAttempt < MAX_START_ATTEMPTS;
      startAttempt += 1
    ) {
      if (claim.state === "completed") {
        return getCompletedApprovalResult(fingerprint, claim.output ?? null);
      }
      if (claim.state === "failed_uncertain") {
        return getUncertainApprovalResult();
      }
      if (
        claim.state === "executing" ||
        claim.state === "in_progress" ||
        claim.state === "ready"
      ) {
        return getPendingApprovalResult();
      }
      if (claim.state !== "claimed" || !claim.leaseToken) {
        throw new Error(
          "Mutation approval claim did not return a usable lease.",
        );
      }

      leaseToken = claim.leaseToken;
      // A cancellation observed before the start request is authoritative:
      // the mutation boundary has not been crossed, so leave the ready lease
      // to expire instead of falsely quarantining it as uncertain.
      throwIfRequestAborted(abortSignal);
      let started: MutationApprovalExecution;
      try {
        // eslint-disable-next-line no-await-in-loop -- A 409 can safely reclaim and retry a ready lease; ambiguous starts still fail closed below.
        started = await startMutationApprovalExecution({
          chatId,
          fingerprint,
          leaseToken,
          toolCallId: approval.toolCallId,
          workspaceSlug,
        });
      } catch (error) {
        if (getErrorStatus(error) === 409) {
          // A conflict is an authoritative response that this lease did not
          // cross the start boundary. Re-claim; never quarantine a definite
          // non-start as an ambiguous mutation.
          // eslint-disable-next-line no-await-in-loop -- Reclaim is bounded by MAX_START_ATTEMPTS and ledgerDeadlineAt.
          claim = await waitForActionableMutationApproval({
            abortSignal,
            approval,
            chatId,
            deadlineAt: ledgerDeadlineAt,
            fingerprint,
            workspaceSlug,
          });
          leaseToken = undefined;
          continue;
        }

        // The start request may have committed even when its response was lost.
        // eslint-disable-next-line no-console -- Preserve protocol diagnostics without logging the approved payload.
        console.error(
          `[chat/route] Could not confirm ${approval.toolName} execution start`,
          getChatErrorDiagnostic(error),
        );
        // eslint-disable-next-line no-await-in-loop -- Start attempts are bounded and an ambiguous transition must be quarantined before returning.
        const failed = await failMutationApprovalSafely({
          approval,
          chatId,
          failureCode: "start_transition_uncertain",
          fingerprint,
          leaseToken,
          workspaceSlug,
        });
        if (failed?.state === "completed") {
          return getCompletedApprovalResult(fingerprint, failed.output ?? null);
        }
        return getUncertainApprovalResult();
      }

      if (started.state === "completed") {
        return getCompletedApprovalResult(fingerprint, started.output ?? null);
      }
      if (started.state === "failed_uncertain") {
        return getUncertainApprovalResult();
      }
      if (started.state !== "started") {
        return getPendingApprovalResult();
      }
      break;
    }

    if (!leaseToken) return getPendingApprovalResult();

    let output: unknown;
    // Once Start is durably acknowledged, a browser disconnect is not proof
    // that the mutation was not dispatched. Use only the execution deadline
    // here so the request does not voluntarily abandon the exact-once path
    // between the start boundary and tool invocation.
    const executionSignal = createExecutionDeadlineSignal(
      APPROVAL_EXECUTION_TIMEOUT_MS,
    );
    try {
      output = await Promise.race([
        executeApprovedMutation({
          abortSignal: executionSignal.signal,
          approval,
          chatId,
          prepared,
          workspaceSlug,
        }),
        waitForAbort(executionSignal.signal),
      ]);
    } catch (error) {
      // eslint-disable-next-line no-console -- Tool exceptions after the start boundary have an ambiguous mutation outcome.
      console.error(
        `[chat/route] Approved ${approval.toolName} execution became uncertain`,
        getChatErrorDiagnostic(error),
      );
      const failed = await failMutationApprovalSafely({
        approval,
        chatId,
        failureCode: "completion_persistence_uncertain",
        fingerprint,
        leaseToken,
        workspaceSlug,
      });
      if (failed?.state === "completed") {
        return getCompletedApprovalResult(fingerprint, failed.output ?? null);
      }
      return getUncertainApprovalResult();
    } finally {
      executionSignal.cleanup();
    }

    try {
      const completed = await completeMutationApprovalExecution({
        chatId,
        fingerprint,
        leaseToken,
        output,
        toolCallId: approval.toolCallId,
        workspaceSlug,
      });
      if (completed.state === "completed") {
        return getCompletedApprovalResult(
          fingerprint,
          completed.output ?? null,
        );
      }
      return getUncertainApprovalResult();
    } catch (error) {
      // Completion can commit while its response is lost. Failure recording is
      // idempotent and returns the completed output when completion won.
      // eslint-disable-next-line no-console -- Preserve protocol diagnostics without logging the approved payload.
      console.error(
        `[chat/route] Could not confirm ${approval.toolName} completion`,
        getChatErrorDiagnostic(error),
      );
      const failed = await failMutationApprovalSafely({
        approval,
        chatId,
        failureCode: "completion_persistence_uncertain",
        fingerprint,
        leaseToken,
        workspaceSlug,
      });
      if (failed?.state === "completed") {
        return getCompletedApprovalResult(fingerprint, failed.output ?? null);
      }
      return getUncertainApprovalResult();
    }
  } catch (error) {
    // eslint-disable-next-line no-console -- Log protocol failures without logging approved payloads.
    console.error(
      `[chat/route] ${approval.toolName} approval ledger failed`,
      getChatErrorDiagnostic(error),
    );
    return {
      denied: false,
      haltsFollowing: true,
      output: getApprovalFailureOutput(
        "Maya could not safely confirm this approved change. Check the workspace before asking Maya to prepare it again.",
      ),
    };
  }
};

export const executeApprovalOnce = ({
  abortSignal,
  approval,
  chatId,
  userId,
  workspaceSlug,
}: {
  abortSignal: AbortSignal;
  approval: MutationToolApproval;
  chatId: string;
  userId: string;
  workspaceSlug: string;
}): Promise<ApprovalExecutionResult> => {
  const cacheKey = JSON.stringify([
    userId,
    workspaceSlug,
    chatId,
    approval.toolCallId,
  ]);
  const fingerprint = getApprovalFingerprint(approval);
  const existing = pendingApprovalExecutions.get(cacheKey);
  if (existing) {
    if (existing.fingerprint === fingerprint) return existing.result;

    return Promise.resolve({
      denied: false,
      haltsFollowing: true,
      output: getApprovalFailureOutput(
        "This approval is stale or no longer matches the prepared change. Ask Maya to prepare it again.",
      ),
    });
  }

  const result = executeApprovalWithLedger({
    abortSignal,
    approval,
    chatId,
    workspaceSlug,
  });
  const pendingExecution = { fingerprint, result };
  pendingApprovalExecutions.set(cacheKey, pendingExecution);
  void result.finally(() => {
    if (pendingApprovalExecutions.get(cacheKey) === pendingExecution) {
      pendingApprovalExecutions.delete(cacheKey);
    }
  });
  return result;
};

export const resetMutationApprovalCacheForTests = () => {
  pendingApprovalExecutions.clear();
};
