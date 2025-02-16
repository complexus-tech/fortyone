"use server";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { teamKeys } from "@/constants/keys";
import type { Team } from "../types";

export const getTeams = async (): Promise<Team[]> => {
  const response = await get<ApiResponse<Team[]>>("teams", {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 10,
      tags: [teamKeys.lists().join("/")],
    },
  });
  return response.data!;
};
