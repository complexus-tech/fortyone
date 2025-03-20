import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { ObjectiveStatus } from "../types";

export const getObjectiveStatuses = async () => {
  const statuses =
    await get<ApiResponse<ObjectiveStatus[]>>("objective-statuses");
  return statuses.data!;
};
