"use server";
import { revalidateTag } from "next/cache";
import { put } from "@/lib/http";
import { workspaceTags } from "@/constants/keys";
import type { ApiResponse, Workspace } from "@/types";

export type UpdateWorkspaceInput = {
  name: string;
};

export const updateWorkspaceAction = async (
  input: UpdateWorkspaceInput,
): Promise<Workspace> => {
  const workspace = await put<UpdateWorkspaceInput, ApiResponse<Workspace>>(
    "",
    input,
  );
  revalidateTag(workspaceTags.lists());
  return workspace.data!;
};
