import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { StoryAssociation, StoryAssociationType } from "../types";

export type UpdateAssociationPayload = {
  fromStoryId: string;
  toStoryId: string;
  type: StoryAssociationType;
};

export const updateAssociationAction = async (
  storyId: string,
  associationId: string,
  payload: UpdateAssociationPayload,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const res = await put<
      UpdateAssociationPayload,
      ApiResponse<StoryAssociation>
    >(`stories/${storyId}/associations/${associationId}`, payload, ctx);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
