import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const updateCollaboratorsAction = async (
  storyId: string,
  collaboratorIds: string[],
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    return await put<{ collaboratorIds: string[] }, ApiResponse<null>>(
      `stories/${storyId}/collaborators`,
      { collaboratorIds },
      { session: session!, workspaceSlug },
    );
  } catch (error) {
    return getApiError(error);
  }
};
