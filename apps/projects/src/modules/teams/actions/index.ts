"use server";

import { post, put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { Team, CreateTeamInput, UpdateTeamInput } from "../types";

export async function createTeam(input: CreateTeamInput) {
  try {
    const team = await post<CreateTeamInput, ApiResponse<Team>>("teams", input);
    return team;
  } catch (error) {
    return getApiError(error);
  }
}

export async function updateTeam(teamId: string, input: UpdateTeamInput) {
  try {
    const team = await put<UpdateTeamInput, ApiResponse<Team>>(
      `teams/${teamId}`,
      input,
    );
    return team;
  } catch (error) {
    return getApiError(error);
  }
}
