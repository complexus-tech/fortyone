import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Sprint } from "../types";

export const getSprint = async (sprintId: string, signal?: AbortSignal) => {
  const response = await get<ApiResponse<Sprint>>(`sprints/${sprintId}`, {
    signal,
  });
  return response.data;
};
