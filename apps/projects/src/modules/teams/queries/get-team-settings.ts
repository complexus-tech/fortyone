import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { TeamSettings } from "../types";

export const getTeamSettings = async (
  teamId: string,
  session: Session,
): Promise<TeamSettings> => {
  const settings = await get<ApiResponse<TeamSettings>>(
    `teams/${teamId}/settings`,
    session,
  );
  return settings.data!;
};
