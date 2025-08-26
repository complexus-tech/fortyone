import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { ActivitiesResponse } from "../types";

export const getKeyResultActivities = async (
  keyResultId: string,
  session: Session,
  page = 1,
  pageSize = 20,
) => {
  const response = await get<ApiResponse<ActivitiesResponse>>(
    `key-results/${keyResultId}/activities?page=${page}&pageSize=${pageSize}`,
    session,
  );
  return response.data!;
};
