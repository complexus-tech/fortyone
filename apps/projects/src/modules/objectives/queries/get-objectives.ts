import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Objective } from "../types";

export const getObjectives = async (session: Session) => {
  const objectives = await get<ApiResponse<Objective[]>>("objectives", session);
  return objectives.data ?? [];
};
