import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { DetailedStory } from "../types";

export const getStory = async (id: string, signal?: AbortSignal) => {
  const response = await get<ApiResponse<DetailedStory>>(`stories/${id}`, {
    signal,
  });
  return response.data!;
};
