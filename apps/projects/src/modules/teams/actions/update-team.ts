"use server";

import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Team } from "@/modules/teams/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export type UpdateTeamInput = {
  name?: string;
  description?: string;
  code?: string;
};

export const updateTeamAction = async (
  id: string,
  input: UpdateTeamInput,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const team = await put<UpdateTeamInput, ApiResponse<Team>>(
      `teams/${id}`,
      input,
      ctx,
    );
    return team;
  } catch (error) {
    return getApiError(error);
  }
};
