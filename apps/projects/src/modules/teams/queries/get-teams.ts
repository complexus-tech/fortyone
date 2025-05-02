import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Team } from "../types";

export const getTeams = async (session: Session): Promise<Team[]> => {
  const response = await get<ApiResponse<Team[]>>("teams", session);
  return response.data!;
};
