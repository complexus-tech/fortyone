"use server";

import { revalidateTag } from "next/cache";
import type { StateCategory } from "@/types/states";
import { patch } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { objectiveTags } from "../../constants";
import type { ObjectiveStatus } from "../../types";

export type UpdateObjectiveStatus = {
  name?: string;
  category?: StateCategory;
  color?: string;
};

export const updateObjectiveStatusAction = async (
  statusId: string,
  payload: UpdateObjectiveStatus,
) => {
  const response = await patch<
    UpdateObjectiveStatus,
    ApiResponse<ObjectiveStatus>
  >(`objective-statuses/${statusId}`, payload);

  revalidateTag(objectiveTags.statuses());

  if (!response.data) {
    throw new Error("Failed to update objective status");
  }

  return response.data;
};
