import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Sprint } from "../types";

export const getSprints = async () => {
  const sprints = await get<ApiResponse<Sprint[]>>("sprints");
  return sprints.data;
};
