"use server";

import type { StateCategory } from "@/types/states";
import { patch } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { ObjectiveStatus } from "../../types";

export type UpdateObjectiveStatus = {
  name?: string;
  category?: StateCategory;
  color?: string;
};

export const updateObjectiveStatusAction = async (
  statusId: string,
  payload: UpdateObjectiveStatus,
) => {
  try {
    const response = await patch<
      UpdateObjectiveStatus,
      ApiResponse<ObjectiveStatus>
    >(`objective-statuses/${statusId}`, payload);
    return response.data;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update objective status");
  }
};
