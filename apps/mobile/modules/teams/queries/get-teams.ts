import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Team } from "../types";

export const getTeams = async (signal?: AbortSignal) => {
  const response = await get<ApiResponse<Team[]>>("teams?joinedOnly=true", {
    signal,
  });
  return response.data!;
};
