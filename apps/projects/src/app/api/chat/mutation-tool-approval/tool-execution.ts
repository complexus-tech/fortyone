import "server-only";

import { AsyncLocalStorage } from "node:async_hooks";
import type { ToolExecutionOptions } from "ai";
import {
  type ApiErrorOutcomeReport,
  installApiErrorOutcomeReporter,
} from "@/utils/api-error-outcome";
import { runWithMayaHttpRequestContext } from "../maya-http-request-context";
import type { MutationToolApproval } from "./approval-fingerprint";
import type { PreparedApprovedMutation } from "./approval-policy";

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

export const getApprovalAbortError = (abortSignal: AbortSignal) =>
  abortSignal.reason instanceof Error
    ? abortSignal.reason
    : new Error("The approval request was aborted.");

export const throwIfRequestAborted = (abortSignal: AbortSignal) => {
  if (abortSignal.aborted) throw getApprovalAbortError(abortSignal);
};

const createTimeoutError = () =>
  Object.assign(
    new Error("The approved change exceeded its execution deadline."),
    {
      code: "approval_execution_timeout",
    },
  );

export const createExecutionDeadlineSignal = (timeoutMs: number) => {
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

export const waitForAbort = (abortSignal: AbortSignal) =>
  new Promise<never>((_, reject) => {
    if (abortSignal.aborted) {
      reject(getApprovalAbortError(abortSignal));
      return;
    }
    abortSignal.addEventListener(
      "abort",
      () => {
        reject(getApprovalAbortError(abortSignal));
      },
      { once: true },
    );
  });

export const isFailedMutationOutput = (output: unknown) => {
  if (!output || typeof output !== "object" || Array.isArray(output)) {
    return false;
  }

  if ("success" in output && output.success === false) return true;
  return (
    "error" in output && output.error !== null && output.error !== undefined
  );
};

export const executeApprovedMutation = async ({
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
