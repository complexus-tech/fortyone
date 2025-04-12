"use server";

import type { StateCategory } from "@/types/states";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { ObjectiveStatus } from "../../types";

export type NewObjectiveStatus = {
  name: string;
  category: StateCategory;
};

export const createObjectiveStatusAction = async (
  newStatus: NewObjectiveStatus,
) => {
  try {
    const response = await post<
      NewObjectiveStatus,
      ApiResponse<ObjectiveStatus>
    >("objective-statuses", newStatus);
    return response;
  } catch (error) {
    return getApiError(error);
  }
};
