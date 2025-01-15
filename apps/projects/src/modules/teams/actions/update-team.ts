"use server";
import { revalidateTag } from "next/cache";
import { put } from "@/lib/http";
import { teamTags } from "@/constants/keys";
import type { ApiResponse } from "@/types";
import type { Team } from "@/modules/teams/types";

export type UpdateTeamInput = {
  name?: string;
  description?: string;
};

export const updateTeamAction = async (
  id: string,
  input: UpdateTeamInput,
): Promise<Team> => {
  const team = await put<UpdateTeamInput, ApiResponse<Team>>(
    `teams/${id}`,
    input,
  );
  revalidateTag(teamTags.detail(id));
  revalidateTag(teamTags.lists());
  return team.data!;
};
