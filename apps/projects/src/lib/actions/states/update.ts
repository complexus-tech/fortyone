"use server";
import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { State } from "@/types/states";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export type UpdateState = {
  name?: string;
  orderIndex?: number;
  isDefault?: boolean;
};

export const updateStateAction = async (
  stateId: string,
  payload: UpdateState,
) => {
  try {
    const session = await auth();
    const res = await put<UpdateState, ApiResponse<State>>(
      `states/${stateId}`,
      payload,
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
