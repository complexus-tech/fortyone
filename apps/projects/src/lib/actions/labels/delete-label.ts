"use server";

import { revalidateTag } from "next/cache";
import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { labelTags } from "@/constants/keys";
import { getApiError } from "@/utils";

export const deleteLabelAction = async (labelId: string) => {
  try {
    await remove<ApiResponse<void>>(`labels/${labelId}`);
    revalidateTag(labelTags.lists());
    return labelId;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete label");
  }
};
