"use server";
import { revalidateTag } from "next/cache";
import { remove } from "@/lib/http";
import { teamTags, statusTags } from "@/constants/keys";
import type { ApiResponse } from "@/types";

export const deleteTeamAction = async (id: string) => {
  await remove<ApiResponse<void>>(`teams/${id}`);
  revalidateTag(teamTags.detail(id));
  revalidateTag(teamTags.lists());
  revalidateTag(statusTags.lists());
};
