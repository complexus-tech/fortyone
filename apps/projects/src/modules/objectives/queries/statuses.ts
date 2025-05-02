import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { ObjectiveStatus } from "../types";

export const getObjectiveStatuses = async (session: Session) => {
  const statuses = await get<ApiResponse<ObjectiveStatus[]>>(
    "objective-statuses",
    session,
  );
  return statuses.data!;
};
