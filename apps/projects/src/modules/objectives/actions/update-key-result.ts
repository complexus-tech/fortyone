"use server";

import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { KeyResultUpdate } from "../types";

export const updateKeyResult = async (
  keyResultId: string,
  objectiveId: string,
  params: KeyResultUpdate,
) => {
  try {
    const res = await put<KeyResultUpdate, ApiResponse<null>>(
      `key-results/${keyResultId}`,
      params,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
