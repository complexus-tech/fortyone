import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import type { KeyResult } from "../types";
import { objectiveTags } from "../constants";

export const getKeyResults = async (objectiveId: string) => {
  const keyResults = await get<ApiResponse<KeyResult[]>>(
    `objectives/${objectiveId}key-results`,
    {
      next: {
        revalidate: DURATION_FROM_SECONDS.MINUTE * 10,
        tags: [objectiveTags.keyResults(objectiveId)],
      },
    },
  );
  return keyResults.data ?? [];
};
