"use server";

import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { State } from "@/types/states";
import { getApiError } from "@/utils";

export type UpdateState = {
  name?: string;
  orderIndex?: number;
};

export const updateStateAction = async (
  stateId: string,
  payload: UpdateState,
) => {
  try {
    const res = await put<UpdateState, ApiResponse<State>>(
      `states/${stateId}`,
      payload,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
