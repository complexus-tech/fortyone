"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { State, StateCategory } from "@/types/states";
import { getApiError } from "@/utils";

export type NewState = {
  name: string;
  category: StateCategory;
  teamId: string;
};

export const createStateAction = async (payload: NewState) => {
  try {
    const state = await post<NewState, ApiResponse<State>>("states", payload);
    return state.data!;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to create state");
  }
};
