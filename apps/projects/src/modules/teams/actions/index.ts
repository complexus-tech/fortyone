"use server";

import { revalidateTag } from "next/cache";
import { post, put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { teamTags, statusTags } from "@/constants/keys";
import { getApiError } from "@/utils";
import type { Team, CreateTeamInput, UpdateTeamInput } from "../types";

export async function createTeam(input: CreateTeamInput) {
  try {
    const team = await post<CreateTeamInput, ApiResponse<Team>>("teams", input);
    revalidateTag(teamTags.lists());
    revalidateTag(statusTags.lists());
    return team;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to create team");
  }
}

export async function updateTeam(
  teamId: string,
  input: UpdateTeamInput,
): Promise<Team> {
  try {
    const team = await put<UpdateTeamInput, ApiResponse<Team>>(
      `teams/${teamId}`,
      input,
    );
    revalidateTag(teamTags.detail(teamId));
    return team.data!;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update team");
  }
}
