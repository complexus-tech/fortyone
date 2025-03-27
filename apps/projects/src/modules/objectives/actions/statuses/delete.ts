"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteObjectiveStatusAction = async (statusId: string) => {
  try {
    const response = await remove<ApiResponse<null>>(
      `objective-statuses/${statusId}`,
    );
    return response;
  } catch (error) {
    return getApiError(error);
  }
};
