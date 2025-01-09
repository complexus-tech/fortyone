import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Objective } from "../types";

export const getObjectives = async () => {
  const objectives = await get<ApiResponse<Objective[]>>("objectives");
  return objectives.data ?? [];
};
