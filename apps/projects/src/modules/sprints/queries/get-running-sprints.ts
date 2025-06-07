import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Sprint } from "../types";

export const getRunningSprints = async (session: Session) => {
  const sprints = await get<ApiResponse<Sprint[]>>("sprints/running", session);
  return sprints.data!;
};
