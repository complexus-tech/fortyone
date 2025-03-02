"use server";
import { revalidateTag } from "next/cache";
import { put } from "@/lib/http";
import { teamTags } from "@/constants/keys";
import type { ApiResponse } from "@/types";
import type { Team } from "@/modules/teams/types";
import { getApiError } from "@/utils";

export type UpdateTeamInput = {
  name?: string;
  description?: string;
};

export const updateTeamAction = async (
  id: string,
  input: UpdateTeamInput,
): Promise<Team> => {
  try {
    const team = await put<UpdateTeamInput, ApiResponse<Team>>(
      `teams/${id}`,
      input,
    );
    revalidateTag(teamTags.detail(id));
    revalidateTag(teamTags.lists());
    return team.data!;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update team");
  }
};
