"use server";

import { get } from "@/lib/http";
import type { ApiResponse, Member } from "@/types";

export const getMembers = async () => {
  const members = await get<ApiResponse<Member[]>>("members");
  return members.data!;
};

export const getTeamMembers = async (teamId: string) => {
  const members = await get<ApiResponse<Member[]>>(`members?teamId=${teamId}`);
  return members.data!;
};
