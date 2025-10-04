import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Sprint } from "../types";

export const getSprint = async (sprintId: string) => {
  const response = await get<ApiResponse<Sprint>>(`sprints/${sprintId}`);
  return response.data;
};
