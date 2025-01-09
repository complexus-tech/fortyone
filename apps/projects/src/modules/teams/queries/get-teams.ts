"use server";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { get } from "@/lib/http";
import type { Team } from "@/modules/teams/types";
import type { ApiResponse } from "@/types";

export const getTeams = async () => {
  const teams = await get<ApiResponse<Team[]>>("teams", {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 10,
      tags: ["teams"],
    },
  });
  return teams.data;
};
