"use server";

import { put } from "@/lib/http";
import type { ApiResponse, Terminology } from "@/types";
import { getApiError } from "@/utils";

export type UpdateTerminology = Partial<Terminology>;

export const updateTerminologyAction = async (payload: UpdateTerminology) => {
  try {
    const res = await put<UpdateTerminology, ApiResponse<Terminology>>(
      `terminology`,
      payload,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
