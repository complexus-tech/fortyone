import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { KeyResult } from "../types";

export const getKeyResults = async (objectiveId: string, session: Session) => {
  const keyResults = await get<ApiResponse<KeyResult[]>>(
    `objectives/${objectiveId}/key-results`,
    session,
  );
  return keyResults.data ?? [];
};
