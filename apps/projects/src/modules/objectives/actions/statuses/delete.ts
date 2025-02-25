"use server";

import { revalidateTag } from "next/cache";
import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { objectiveTags } from "../../constants";

export const deleteObjectiveStatusAction = async (statusId: string) => {
  await remove<ApiResponse<void>>(`objective-statuses/${statusId}`);
  revalidateTag(objectiveTags.statuses());
  return true;
};
