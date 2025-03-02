"use server";

import { revalidateTag } from "next/cache";
import { post } from "@/lib/http";
import type { ApiResponse, Label } from "@/types";
import { labelTags } from "@/constants/keys";
import { getApiError } from "@/utils";

export type NewLabel = {
  name: string;
  color: string;
  teamId?: string;
};

export const createLabelAction = async (newLabel: NewLabel) => {
  try {
    const label = await post<NewLabel, ApiResponse<Label>>("labels", newLabel);
    revalidateTag(labelTags.lists());
    return label.data!;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to create label");
  }
};
