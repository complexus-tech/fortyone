import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Link } from "@/types/link";

export const getLinks = async (storyId: string) => {
  const response = await get<ApiResponse<Link[]>>(`stories/${storyId}/links`);
  return response.data ?? [];
};
