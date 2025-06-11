import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { State } from "@/types/states";

export const getTeamStatuses = async (teamId: string, session: Session) => {
  const statuses = await get<ApiResponse<State[]>>(
    `states?teamId=${teamId}`,
    session,
  );
  return statuses.data!;
};
