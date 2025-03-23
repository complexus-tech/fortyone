import { put } from "@/lib/http";
import type { ApiResponse, Workspace } from "@/types";
import { getApiError } from "@/utils";

export type UpdateWorkspaceInput = {
  name: string;
};

export const updateWorkspaceAction = async (input: UpdateWorkspaceInput) => {
  try {
    const workspace = await put<UpdateWorkspaceInput, ApiResponse<Workspace>>(
      "",
      input,
    );
    return workspace;
  } catch (error) {
    return getApiError(error);
  }
};
