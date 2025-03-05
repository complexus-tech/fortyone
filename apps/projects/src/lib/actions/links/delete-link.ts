"use server";

import { remove } from "@/lib/http";
import { getApiError } from "@/utils";

export const deleteLinkAction = async (linkId: string) => {
  try {
    await remove(`links/${linkId}`);
    return linkId;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete link");
  }
};
