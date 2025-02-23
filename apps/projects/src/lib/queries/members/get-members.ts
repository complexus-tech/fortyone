"use server";
import { memberTags } from "@/constants/keys";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { get } from "@/lib/http";
import type { ApiResponse, Member } from "@/types";

export const getMembers = async () => {
  const members = await get<ApiResponse<Member[]>>("members", {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 10,
      tags: [memberTags.lists()],
    },
  });
  return members.data!;
};

export const getTeamMembers = async (teamId: string) => {
  const members = await get<ApiResponse<Member[]>>(`members?teamId=${teamId}`, {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 10,
      tags: [memberTags.team(teamId)],
    },
  });
  return members.data!;
};
