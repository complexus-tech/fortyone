"use server";
import { revalidateTag } from "next/cache";
import { put } from "@/lib/http";
import { workspaceTags } from "@/constants/keys";
import type { ApiResponse, Workspace } from "@/types";
import { getApiError } from "@/utils";

export type UpdateWorkspaceInput = {
  name: string;
};

export const updateWorkspaceAction = async (
  input: UpdateWorkspaceInput,
): Promise<Workspace> => {
  try {
    const workspace = await put<UpdateWorkspaceInput, ApiResponse<Workspace>>(
      "",
      input,
    );
    revalidateTag(workspaceTags.lists());
    return workspace.data!;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update workspace");
  }
};
