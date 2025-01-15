"use server";
import { revalidateTag } from "next/cache";
import { put } from "@/lib/http";
import { workspaceTags } from "@/constants/keys";
import type { ApiResponse, Workspace } from "@/types";

export type UpdateWorkspaceInput = {
  name?: string;
  description?: string;
};

export const updateWorkspaceAction = async (
  id: string,
  input: UpdateWorkspaceInput,
): Promise<Workspace> => {
  const workspace = await put<UpdateWorkspaceInput, ApiResponse<Workspace>>(
    `workspaces/${id}`,
    input,
  );
  revalidateTag(workspaceTags.detail(id));
  revalidateTag(workspaceTags.lists());
  return workspace.data!;
};
