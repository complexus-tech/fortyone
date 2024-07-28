import { get } from "@/lib/http";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { TAGS } from "@/constants/tags";
import { auth } from "@/auth";
import { Team } from "@/modules/teams/types";

export const getTeams = async () => {
  const session = await auth();
  const teams = await get<Team[]>(`/teams`, {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 30,
      tags: [TAGS.teams],
    },
  });
  return teams;
};
