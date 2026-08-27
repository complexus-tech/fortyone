import { ApiError } from "api-client";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { StoryDeletionOutcomeUncertainError } from "./story-deletion-error";

const isDefiniteStoryDeletionFailure = (error: unknown) =>
  error instanceof ApiError && error.status >= 400 && error.status < 500;

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
