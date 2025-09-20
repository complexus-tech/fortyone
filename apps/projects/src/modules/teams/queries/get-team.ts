import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Team } from "@/modules/teams/types";
import { getApiError } from "@/utils";

export const getTeam = async (id: string, session: Session) => {
  try {
    const data = await get<ApiResponse<Team>>(`teams/${id}`, session);
    return data;
  } catch (error) {
    return getApiError(error);
  }
};
