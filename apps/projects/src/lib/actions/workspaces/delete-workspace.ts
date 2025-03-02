"use server";
import { revalidateTag } from "next/cache";
import { remove } from "@/lib/http";
import { workspaceTags } from "@/constants/keys";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteWorkspaceAction = async (id: string): Promise<void> => {
  try {
    await remove<ApiResponse<void>>(`workspaces/${id}`);
    revalidateTag(workspaceTags.lists());
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete workspace");
  }
};
