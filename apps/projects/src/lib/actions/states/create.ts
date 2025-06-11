"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { State, StateCategory } from "@/types/states";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export type NewState = {
  name: string;
  category: StateCategory;
  teamId: string;
  color: string;
};

export const createStateAction = async (payload: NewState) => {
  try {
    const session = await auth();
    const state = await post<NewState, ApiResponse<State>>(
      "states",
      payload,
      session!,
    );
    return state;
  } catch (error) {
    return getApiError(error);
  }
};
