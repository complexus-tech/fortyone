"use server";

import { put } from "@/lib/http";
import type { ApiResponse, Label } from "@/types";
import { getApiError } from "@/utils";

export type UpdateLabel = {
  name: string;
  color: string;
};

export const editLabelAction = async (
  labelId: string,
  updates: UpdateLabel,
) => {
  try {
    const label = await put<UpdateLabel, ApiResponse<Label>>(
      `labels/${labelId}`,
      updates,
    );
    return label.data!;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to edit label");
  }
};
