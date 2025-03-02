"use server";

import { revalidateTag } from "next/cache";
import { put } from "@/lib/http";
import type { ApiResponse, Label } from "@/types";
import { labelTags } from "@/constants/keys";
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
    revalidateTag(labelTags.lists());
    return label.data!;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to edit label");
  }
};
