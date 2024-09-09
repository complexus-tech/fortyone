"use server";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { get } from "@/lib/http";
import { Team } from "@/modules/teams/types";

export const getTeams = async () => {
  const teams = await get<Team[]>("teams", {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 10,
      tags: ["teams"],
    },
  });
  return teams;
};
