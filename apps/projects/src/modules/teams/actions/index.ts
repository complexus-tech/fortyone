"use server";
import { post, put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { Team, CreateTeamInput, UpdateTeamInput } from "../types";

export async function createTeam(input: CreateTeamInput) {
  try {
    const session = await auth();
    const team = await post<CreateTeamInput, ApiResponse<Team>>(
      "teams",
      input,
      session!,
    );
    return team;
  } catch (error) {
    return getApiError(error);
  }
}

export async function updateTeam(teamId: string, input: UpdateTeamInput) {
  try {
    const session = await auth();
    const team = await put<UpdateTeamInput, ApiResponse<Team>>(
      `teams/${teamId}`,
      input,
      session!,
    );
    return team;
  } catch (error) {
    return getApiError(error);
  }
}

export { updateSprintSettingsAction } from "./update-sprint-settings";
export { updateStoryAutomationSettingsAction } from "./update-story-automation-settings";
