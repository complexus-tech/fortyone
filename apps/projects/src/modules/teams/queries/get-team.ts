"use server";
import { get } from "@/lib/http";
import { teamTags } from "@/constants/keys";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import type { ApiResponse } from "@/types";
import type { Team } from "@/modules/teams/types";

export const getTeam = async (id: string): Promise<Team> => {
  const team = await get<ApiResponse<Team>>(`teams/${id}`, {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 30,
      tags: [teamTags.detail(id)],
    },
  });
  return team.data!;
};
