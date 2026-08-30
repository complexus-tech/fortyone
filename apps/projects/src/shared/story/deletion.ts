import { ApiError } from "api-client";
import type { ApiResponse } from "@/types/api-response";
import { getApiError } from "@/utils/api-error";

const UNCERTAIN_STORY_DELETION_MESSAGE =
  "The story deletion request did not return a definitive result.";

const isDefiniteStoryDeletionFailure = (error: unknown) =>
  error instanceof ApiError && error.status >= 400 && error.status < 500;

/**
 * Signals that a story-deletion request crossed the mutation boundary but its
 * response could not prove whether the server committed the deletion. Approved
 * tool execution must let this reach the mutation ledger for reconciliation.
 */
export class StoryDeletionOutcomeUncertainError extends Error {
  constructor(cause: unknown) {
    super(UNCERTAIN_STORY_DELETION_MESSAGE, { cause });
    this.name = "StoryDeletionOutcomeUncertainError";
  }
}

export const isStoryDeletionOutcomeUncertainError = (
  error: unknown,
): error is StoryDeletionOutcomeUncertainError =>
  error instanceof StoryDeletionOutcomeUncertainError;

const getUncertainStoryDeletionError = (...causes: unknown[]) =>
  new StoryDeletionOutcomeUncertainError(
    causes.length === 1
      ? causes[0]
      : new AggregateError(
          causes,
          "Story deletion remained uncertain after its idempotent retry.",
        ),
  );

type StoryDeletionRequestOptions<Result> = {
  request: () => Promise<Result>;
  retryUncertain: boolean;
};

/**
 * Executes an exact deletion request. Soft deletion is desired-state
 * idempotent, so one bounded retry can recover a committed request whose
 * response was lost. A second ambiguous result remains quarantinable.
 */
export const executeStoryDeletionRequest = async <Result>({
  request,
  retryUncertain,
}: StoryDeletionRequestOptions<Result>): Promise<
  Result | ApiResponse<null>
> => {
  try {
    return await request();
  } catch (error) {
    if (isDefiniteStoryDeletionFailure(error)) return getApiError(error);
    if (!retryUncertain) throw getUncertainStoryDeletionError(error);

    try {
      return await request();
    } catch (retryError) {
      // Even a definite response to the retry cannot disprove that the first
      // request committed before its response was lost.
      throw getUncertainStoryDeletionError(error, retryError);
    }
  }
};
