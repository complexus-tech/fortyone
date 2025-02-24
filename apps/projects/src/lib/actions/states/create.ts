"use server";

import { revalidateTag } from "next/cache";
import { statusTags } from "@/constants/keys";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { State, StateCategory } from "@/types/states";

export type NewState = {
  name: string;
  category: StateCategory;
  teamId: string;
};

export const createStateAction = async (payload: NewState) => {
  const state = await post<NewState, ApiResponse<State>>("states", payload);
  revalidateTag(statusTags.lists());
  return state.data!;
};
