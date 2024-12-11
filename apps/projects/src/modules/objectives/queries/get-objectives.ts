import { get } from "@/lib/http";
import { Objective } from "../types";
import { ApiResponse } from "@/types";

export const getObjectives = async () => {
  const objectives = await get<ApiResponse<Objective[]>>("objectives");
  return objectives?.data ?? [];
};
