import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Team } from "../types";

export const getTeams = async () => {
  const response = await get<ApiResponse<Team[]>>("teams");
  return response.data!;
};
