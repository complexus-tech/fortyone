"use server";

import { post } from "@/lib/http";
import type { ApiResponse, Label } from "@/types";
import { getApiError } from "@/utils";

export type NewLabel = {
  name: string;
  color: string;
  teamId?: string;
};

export const createLabelAction = async (newLabel: NewLabel) => {
  try {
    const label = await post<NewLabel, ApiResponse<Label>>("labels", newLabel);
    return label;
  } catch (error) {
    return getApiError(error);
  }
};
