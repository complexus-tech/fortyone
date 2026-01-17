"use server";
import { post, put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { Team, CreateTeamInput, UpdateTeamInput } from "../types";

export async function createTeam(input: CreateTeamInput, workspaceSlug: string) {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const team = await post<CreateTeamInput, ApiResponse<Team>>("teams", input, ctx);
    return team;
  } catch (error) {
    return getApiError(error);
  }
}

export async function updateTeam(
  teamId: string,
  input: UpdateTeamInput,
  workspaceSlug: string,
) {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const team = await put<UpdateTeamInput, ApiResponse<Team>>(
      `teams/${teamId}`,
      input,
      ctx,
    );
    return team;
  } catch (error) {
    return getApiError(error);
  }
}
