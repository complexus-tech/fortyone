import { get } from "@/lib/http";
import type { ApiResponse, Member } from "@/types";

export const getMembers = async (signal?: AbortSignal) => {
  const response = await get<ApiResponse<Member[]>>("members", { signal });
  return response.data!;
};

export const getTeamMembers = async (teamId: string, signal?: AbortSignal) => {
  const response = await get<ApiResponse<Member[]>>(
    `members?teamId=${teamId}`,
    { signal },
  );
  return response.data!;
};
