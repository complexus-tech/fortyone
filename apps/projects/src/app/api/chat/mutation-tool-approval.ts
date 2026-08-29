import { createHash } from "node:crypto";
import { AsyncLocalStorage } from "node:async_hooks";
import type { FlexibleSchema, ToolExecutionOptions, ToolSet } from "ai";
import { createUIMessageStream, createUIMessageStreamResponse } from "ai";
import { validateToolInputWithStrictNullNormalization } from "@/lib/ai/model-tools";
import { tools } from "@/lib/ai/tools";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import {
  type MayaToolName,
  requiresMutationApproval,
  toApprovedMutationInput,
} from "@/lib/ai/tool-policy";
import {
  claimMutationApprovalExecution,
  completeMutationApprovalExecution,
  failMutationApprovalExecution,
  startMutationApprovalExecution,
} from "@/modules/ai-chats/actions/mutation-approval-execution";
import { recoverMutationApprovalOutput } from "@/modules/ai-chats/actions/message-write";
import type {
  MutationApprovalExecution,
  MutationApprovalFailureCode,
} from "@/modules/ai-chats/actions/mutation-approval-execution";
import {
  type ApiErrorOutcomeReport,
  installApiErrorOutcomeReporter,
} from "@/utils/api-error-outcome";
import { beginChatWrite, saveChat } from "./save-chat";
import { getChatErrorDiagnostic } from "./chat-errors";
import { runWithMayaHttpRequestContext } from "./maya-http-request-context";

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
const SKIPPED_APPROVAL_OUTPUT_MESSAGE =
  "Maya did not run this approved change because an earlier approved change was unresolved. Review the earlier result, then ask Maya to prepare this change again.";

type MutationToolApproval = {
  approved: boolean;
  input: unknown;
  toolCallId: string;
  toolName: MayaToolName;
};

type ApprovalExecutionResult =
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

type PreparedApprovedMutation = {
  execute: NonNullable<ToolSet[string]["execute"]>;
  input: unknown;
};

const pendingApprovalExecutions = new Map<string, PendingApprovalExecution>();

type ApiErrorOutcomeExecution = {
  uncertainFailureObserved: boolean;
};

const API_ERROR_OUTCOME_STORAGE_KEY = Symbol.for(
  "fortyone.maya-api-error-outcome-storage",
);

type GlobalWithApiErrorOutcomeStorage = typeof globalThis & {
  [API_ERROR_OUTCOME_STORAGE_KEY]?: AsyncLocalStorage<ApiErrorOutcomeExecution>;
};

const apiErrorOutcomeStorage = (() => {
  const globalState = globalThis as GlobalWithApiErrorOutcomeStorage;
  const existing = globalState[API_ERROR_OUTCOME_STORAGE_KEY];
  if (existing) return existing;

  const storage = new AsyncLocalStorage<ApiErrorOutcomeExecution>();
  globalState[API_ERROR_OUTCOME_STORAGE_KEY] = storage;
  return storage;
})();

installApiErrorOutcomeReporter((report: ApiErrorOutcomeReport) => {
  if (report.certainty !== "uncertain") return;
  const execution = apiErrorOutcomeStorage.getStore();
  if (execution) execution.uncertainFailureObserved = true;
});

const isRegisteredToolName = (toolName: string): toolName is MayaToolName =>
  Object.hasOwn(tools, toolName);

const getMutationToolApprovals = (
  messages: MayaUIMessage[],
): MutationToolApproval[] => {
  const lastMessage = messages.at(-1);
  if (lastMessage?.role !== "assistant") return [];

  const approvals: MutationToolApproval[] = [];
  const seenToolCallIds = new Set<string>();

  for (const part of lastMessage.parts) {
    if (!part.type.startsWith("tool-") || !("state" in part)) continue;
    if (part.state !== "approval-responded" || !("approval" in part)) {
      continue;
    }
    if (!("input" in part) || !("toolCallId" in part)) continue;
    if (typeof part.toolCallId !== "string") continue;
    if (seenToolCallIds.has(part.toolCallId)) continue;

    const toolName = part.type.slice("tool-".length);
    const approved = part.approval.approved;
    if (
      typeof approved !== "boolean" ||
      !isRegisteredToolName(toolName) ||
      !requiresMutationApproval(toolName, part.input)
    ) {
      continue;
    }

    seenToolCallIds.add(part.toolCallId);
    approvals.push({
      approved,
      input: part.input,
      toolCallId: part.toolCallId,
      toolName,
    });
  }

  return approvals;
};

const normalizeForFingerprint = (value: unknown): unknown => {
  if (Array.isArray(value)) return value.map(normalizeForFingerprint);
  if (!value || typeof value !== "object") return value;

  return Object.fromEntries(
    Object.entries(value)
      .sort(([left], [right]) => left.localeCompare(right))
      .map(([key, item]) => [key, normalizeForFingerprint(item)]),
  );
};

const getFingerprint = (value: unknown) =>
  createHash("sha256")
    .update(JSON.stringify(normalizeForFingerprint(value)))
    .digest("hex");

const getApprovalFingerprint = (approval: MutationToolApproval) =>
  getFingerprint({
    approved: approval.approved,
    input: approval.input,
    toolName: approval.toolName,
  });

const getPreparedApprovalFingerprint = (
  approval: MutationToolApproval,
  input: unknown,
) =>
  getFingerprint({
    approved: true,
    input,
    toolName: approval.toolName,
  });

const getApprovalFailureOutput = (error: string) => ({
  error,
  success: false,
});

const getAbortError = (abortSignal: AbortSignal) =>
  abortSignal.reason instanceof Error
    ? abortSignal.reason
    : new Error("The approval request was aborted.");

const throwIfRequestAborted = (abortSignal: AbortSignal) => {
  if (abortSignal.aborted) throw getAbortError(abortSignal);
};

const createTimeoutError = () =>
  Object.assign(
    new Error("The approved change exceeded its execution deadline."),
    {
      code: "approval_execution_timeout",
    },
  );

const createExecutionDeadlineSignal = (timeoutMs: number) => {
  const controller = new AbortController();
  const timeout = setTimeout(() => {
    controller.abort(createTimeoutError());
  }, timeoutMs);

  return {
    cleanup: () => {
      clearTimeout(timeout);
    },
    signal: controller.signal,
  };
};

const waitForAbort = (abortSignal: AbortSignal) =>
  new Promise<never>((_, reject) => {
    if (abortSignal.aborted) {
      reject(getAbortError(abortSignal));
      return;
    }
    abortSignal.addEventListener(
      "abort",
      () => {
        reject(getAbortError(abortSignal));
      },
      { once: true },
    );
  });

const getErrorStatus = (error: unknown) => {
  if (!error || typeof error !== "object" || !("status" in error)) {
    return undefined;
  }

  return typeof error.status === "number" ? error.status : undefined;
};

const validateToolInput = async (toolName: MayaToolName, input: unknown) => {
  const registeredTool = tools[toolName];
  const validation = await validateToolInputWithStrictNullNormalization(
    registeredTool.inputSchema as FlexibleSchema<unknown>,
    input,
  );
  if (!validation.success) throw validation.error;
  if (!requiresMutationApproval(toolName, validation.value)) {
    throw new Error(`The ${toolName} call is not a mutation.`);
  }

  return validation.value;
};

const waitForApprovalPersistence = (
  delayMs: number,
  abortSignal: AbortSignal,
) =>
  new Promise<void>((resolve, reject) => {
    if (abortSignal.aborted) {
      reject(getAbortError(abortSignal));
      return;
    }

    const timeout = setTimeout(() => {
      abortSignal.removeEventListener("abort", onAbort);
      resolve();
    }, delayMs);
    const onAbort = () => {
      clearTimeout(timeout);
      reject(getAbortError(abortSignal));
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

const prepareApprovedMutation = async (
  approval: MutationToolApproval,
): Promise<PreparedApprovedMutation> => {
  const registeredTool = tools[approval.toolName] as ToolSet[string];
  if (!registeredTool.execute) {
    throw new Error(`The ${approval.toolName} tool cannot be executed.`);
  }

  const validatedInput = await validateToolInput(
    approval.toolName,
    approval.input,
  );
  return {
    execute: registeredTool.execute,
    input: toApprovedMutationInput(approval.toolName, validatedInput),
  };
};

const executeApprovedMutation = async ({
  abortSignal,
  approval,
  chatId,
  prepared,
  workspaceSlug,
}: {
  abortSignal: AbortSignal;
  approval: MutationToolApproval;
  chatId: string;
  prepared: PreparedApprovedMutation;
  workspaceSlug: string;
}): Promise<unknown> => {
  if (abortSignal.aborted) {
    throw abortSignal.reason ?? new Error("The approval request was aborted.");
  }

  const options: ToolExecutionOptions = {
    abortSignal,
    experimental_context: { chatId, workspaceSlug },
    messages: [],
    toolCallId: approval.toolCallId,
  };
  const execution: ApiErrorOutcomeExecution = {
    uncertainFailureObserved: false,
  };
  const output = await apiErrorOutcomeStorage.run(execution, () =>
    runWithMayaHttpRequestContext(abortSignal, () =>
      prepared.execute(prepared.input, options),
    ),
  );

  if (execution.uncertainFailureObserved && isFailedMutationOutput(output)) {
    throw new Error(
      "The approved mutation returned an unconfirmed API failure after execution started.",
    );
  }

  return output;
};

const isFailedMutationOutput = (output: unknown) => {
  if (!output || typeof output !== "object" || Array.isArray(output)) {
    return false;
  }

  if ("success" in output && output.success === false) return true;
  return (
    "error" in output && output.error !== null && output.error !== undefined
  );
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

const executeApprovalOnce = ({
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

export const createMutationToolApprovalResponse = ({
  abortSignal,
  chatId,
  messageId,
  messages,
  userId,
  workspaceSlug,
}: {
  abortSignal: AbortSignal;
  chatId: string;
  messageId?: string;
  messages: MayaUIMessage[];
  userId?: string;
  workspaceSlug: string;
}): Promise<Response> | Response | undefined => {
  const submittedApprovals = getMutationToolApprovals(messages);
  if (submittedApprovals.length === 0) return undefined;
  if (!userId) return new Response("Unauthorized", { status: 401 });

  return (async () => {
    // Reserve and validate the persisted transition before creating the
    // persistence stream. If the server repaired durable receipts, its
    // request-safe transcript becomes the base for both streaming and final
    // CAS persistence instead of letting a stale browser overwrite it.
    const reservation = await beginChatWrite({
      id: chatId,
      messageId,
      messages,
      operation: "approval",
      workspaceSlug,
    });
    const canonicalMessages = reservation.messages ?? messages;
    const approvals = getMutationToolApprovals(canonicalMessages);
    const recoverableApprovals = new Map<string, string>();

    const stream = createUIMessageStream<MayaUIMessage>({
      execute: async ({ writer }) => {
        // The server may have repaired terminal receipts while reserving this
        // write. Re-emit them so the browser converges immediately instead of
        // waiting for a refresh; replaying an identical terminal chunk is
        // idempotent in AI SDK's UI-message state machine.
        const canonicalLastMessage = canonicalMessages.at(-1);
        if (canonicalLastMessage?.role === "assistant") {
          for (const part of canonicalLastMessage.parts) {
            if (
              !part.type.startsWith("tool-") ||
              !("state" in part) ||
              !("toolCallId" in part) ||
              typeof part.toolCallId !== "string"
            ) {
              continue;
            }
            if (part.state === "output-denied") {
              writer.write({
                toolCallId: part.toolCallId,
                type: "tool-output-denied",
              });
            } else if (part.state === "output-available" && "output" in part) {
              writer.write({
                output: part.output,
                toolCallId: part.toolCallId,
                type: "tool-output-available",
              });
            }
          }
        }

        let haltFollowingApprovedMutations = false;
        for (const approval of approvals) {
          if (!approval.approved) {
            writer.write({
              toolCallId: approval.toolCallId,
              type: "tool-output-denied",
            });
            continue;
          }

          if (haltFollowingApprovedMutations) {
            writer.write({
              output: getApprovalFailureOutput(SKIPPED_APPROVAL_OUTPUT_MESSAGE),
              toolCallId: approval.toolCallId,
              type: "tool-output-available",
            });
            continue;
          }

          // eslint-disable-next-line no-await-in-loop -- Mutations must preserve the approved order and never execute concurrently.
          const result = await executeApprovalOnce({
            abortSignal,
            approval,
            chatId,
            userId,
            workspaceSlug,
          });

          if (result.denied) {
            writer.write({
              toolCallId: approval.toolCallId,
              type: "tool-output-denied",
            });
            continue;
          }

          if (result.durableFingerprint) {
            recoverableApprovals.set(
              approval.toolCallId,
              result.durableFingerprint,
            );
          }
          haltFollowingApprovedMutations = result.haltsFollowing;

          writer.write({
            output: result.output,
            toolCallId: approval.toolCallId,
            type: "tool-output-available",
          });
        }
      },
      onFinish: async ({ messages: finishedMessages }) => {
        let finalizationError: unknown;
        let applied = false;
        try {
          const result = await saveChat({
            id: chatId,
            messages: finishedMessages,
            reservation,
            workspaceSlug,
          });
          applied = result.applied;
        } catch (error) {
          finalizationError = error;
        }
        if (applied) return;

        if (recoverableApprovals.size === 0) {
          if (finalizationError) {
            throw finalizationError instanceof Error
              ? finalizationError
              : new Error("Maya transcript finalization failed.", {
                  cause: finalizationError,
                });
          }
          return;
        }

        try {
          for (const [toolCallId, fingerprint] of recoverableApprovals) {
            // eslint-disable-next-line no-await-in-loop -- Each targeted merge preserves the exact durable approval order and surrounding transcript.
            await recoverMutationApprovalOutput({
              chatId,
              fingerprint,
              toolCallId,
              workspaceSlug,
            });
          }
        } catch (recoveryError) {
          // eslint-disable-next-line no-console -- Recovery diagnostics omit message and tool payloads.
          console.error(
            "[chat/route] Could not project durable approval output into the transcript",
            getChatErrorDiagnostic(recoveryError),
          );
          throw finalizationError ?? recoveryError;
        }
      },
      originalMessages: canonicalMessages,
    });

    return createUIMessageStreamResponse({ stream });
  })();
};
