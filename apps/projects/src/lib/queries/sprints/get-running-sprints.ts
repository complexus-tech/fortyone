import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { Sprint } from "@/modules/sprints/types";
import type { ApiResponse } from "@/types";

export const getRunningSprints = async (session: Session) => {
  const sprints = await get<ApiResponse<Sprint[]>>("sprints", session);
  return sprints.data!;
};
