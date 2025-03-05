import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { KeyResult } from "../types";

export const getKeyResults = async (objectiveId: string) => {
  const keyResults = await get<ApiResponse<KeyResult[]>>(
    `objectives/${objectiveId}/key-results`,
  );
  return keyResults.data ?? [];
};
