import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { SprintDetails } from "../types";

export const getSprint = async (sprintId: string, session: Session) => {
  const sprint = await get<ApiResponse<SprintDetails>>(
    `sprints/${sprintId}`,
    session,
  );
  return sprint.data;
};
