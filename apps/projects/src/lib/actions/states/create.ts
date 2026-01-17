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

export const createStateAction = async (
  payload: NewState,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const state = await post<NewState, ApiResponse<State>>("states", payload, ctx);
    return state;
  } catch (error) {
    return getApiError(error);
  }
};
