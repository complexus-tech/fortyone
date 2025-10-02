import { get } from "@/lib/http";
import type { ApiResponse, Member } from "@/types";

export const getMembers = async () => {
  const response = await get<ApiResponse<Member[]>>("members");
  return response.data!;
};

export const getTeamMembers = async (teamId: string) => {
  const response = await get<ApiResponse<Member[]>>(`members?teamId=${teamId}`);
  return response.data!;
};
