import { ApiError } from "api-client";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { DetailedStory, NewStory } from "../types";
import { StoryCreationOutcomeUncertainError } from "./story-creation-error";

const isDefiniteStoryCreationFailure = (error: unknown) =>
  error instanceof ApiError && error.status >= 400 && error.status < 500;

const getUncertainStoryCreationError = (...causes: unknown[]) =>
  new StoryCreationOutcomeUncertainError(
    causes.length === 1
      ? causes[0]
      : new AggregateError(
          causes,
          "Story creation remained uncertain after its idempotent retry.",
        ),
  );

export const createStoryAction = async (
  newStory: NewStory,
  workspaceSlug: string,
) => {
  const session = await auth();
  const ctx = { session: session!, workspaceSlug };
  const createStory = () =>
    post<NewStory, ApiResponse<DetailedStory>>("stories", newStory, ctx);

  try {
    return await createStory();
  } catch (error) {
    if (isDefiniteStoryCreationFailure(error)) return getApiError(error);
    if (!newStory.idempotencyKey?.trim()) {
      throw getUncertainStoryCreationError(error);
    }

    try {
      return await createStory();
    } catch (retryError) {
      // Even a definite response to the retry cannot disprove that the first
      // request committed before its response was lost.
      throw getUncertainStoryCreationError(error, retryError);
    }
  }
};
