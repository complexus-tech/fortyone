import { get } from "@/lib/http";
import { Sprint } from "../types";
import { ApiResponse } from "@/types";

export const getSprints = async () => {
  const sprints = await get<ApiResponse<Sprint[]>>("sprints");
  return sprints?.data;
};
