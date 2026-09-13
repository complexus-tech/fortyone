import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Link } from "@/types/link";

export const getLinks = async (storyId: string, signal?: AbortSignal) => {
  const response = await get<ApiResponse<Link[]>>(`stories/${storyId}/links`, {
    signal,
  });
  return response.data ?? [];
};
