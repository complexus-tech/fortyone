"use server";

import { revalidateTag } from "next/cache";
import { statusTags } from "@/constants/keys";
import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { State } from "@/types/states";

export type UpdateState = {
  name?: string;
  orderIndex?: number;
};

export const updateStateAction = async (
  stateId: string,
  payload: UpdateState,
) => {
  const _ = await put<UpdateState, ApiResponse<State>>(
    `states/${stateId}`,
    payload,
  );
  revalidateTag(statusTags.lists());
  return stateId;
};
