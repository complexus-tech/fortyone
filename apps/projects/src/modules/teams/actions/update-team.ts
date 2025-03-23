import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Team } from "@/modules/teams/types";
import { getApiError } from "@/utils";

export type UpdateTeamInput = {
  name?: string;
  description?: string;
};

export const updateTeamAction = async (id: string, input: UpdateTeamInput) => {
  try {
    const team = await put<UpdateTeamInput, ApiResponse<Team>>(
      `teams/${id}`,
      input,
    );
    return team;
  } catch (error) {
    return getApiError(error);
  }
};
