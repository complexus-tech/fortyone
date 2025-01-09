"use server";
import "server-only";

import { auth } from "@/auth";
import { get } from "@/lib/http";
import { TAGS } from "@/constants/tags";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import type { ApiResponse } from "@/types";
import type { Epic } from "../types";

export const getTeamEpics = async () => {
  const epics = await get<ApiResponse<Epic[]>>("/epics", {
    next: {
      tags: [TAGS.epics],
      revalidate: Number(DURATION_FROM_SECONDS.HOUR),
    },
  });

  return epics.data;
};
