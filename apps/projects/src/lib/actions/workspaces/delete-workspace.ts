"use server";
import { revalidateTag } from "next/cache";
import { remove } from "@/lib/http";
import { workspaceTags } from "@/constants/keys";
import type { ApiResponse } from "@/types";

export const deleteWorkspaceAction = async (id: string): Promise<void> => {
  await remove<ApiResponse<void>>(`workspaces/${id}`);
  revalidateTag(workspaceTags.detail());
  revalidateTag(workspaceTags.lists());
};
