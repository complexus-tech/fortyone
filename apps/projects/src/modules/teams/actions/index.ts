"use server";

import { revalidateTag } from "next/cache";
import { post, put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { teamTags } from "@/constants/keys";
import { getApiError } from "@/utils";
import type { Team, CreateTeamInput, UpdateTeamInput } from "../types";

export async function createTeam(input: CreateTeamInput) {
  try {
    const team = await post<CreateTeamInput, ApiResponse<Team>>("teams", input);
    revalidateTag(teamTags.lists());
    return team;
  } catch (error) {
    return getApiError(error);
  }
}

export async function updateTeam(
  teamId: string,
  input: UpdateTeamInput,
): Promise<Team> {
  const team = await put<UpdateTeamInput, ApiResponse<Team>>(
    `teams/${teamId}`,
    input,
  );
  revalidateTag(teamTags.detail(teamId));
  revalidateTag(teamTags.lists());
  return team.data!;
}

export async function addTeamMember(
  teamId: string,
  userId: string,
  role: "admin" | "member",
): Promise<void> {
  await post<{ userId: string; role: string }, ApiResponse<void>>(
    `teams/${teamId}/members`,
    { userId, role },
  );
  revalidateTag(teamTags.lists());
}
