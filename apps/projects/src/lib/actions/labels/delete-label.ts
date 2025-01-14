"use server";

import { revalidateTag } from "next/cache";
import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { labelTags } from "@/constants/keys";

export const deleteLabelAction = async (labelId: string) => {
  await remove<ApiResponse<void>>(`labels/${labelId}`);
  revalidateTag(labelTags.lists());
};
