"use server";
import { revalidateTag } from "next/cache";
import { remove } from "@/lib/http";
import { teamTags, statusTags } from "@/constants/keys";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteTeamAction = async (id: string) => {
  try {
    await remove<ApiResponse<void>>(`teams/${id}`);
    revalidateTag(teamTags.detail(id));
    revalidateTag(teamTags.lists());
    revalidateTag(statusTags.lists());
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete team");
  }
};
