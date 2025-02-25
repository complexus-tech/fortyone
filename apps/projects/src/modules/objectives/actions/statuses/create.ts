"use server";

import { revalidateTag } from "next/cache";
import type { StateCategory } from "@/types/states";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { objectiveTags } from "../../constants";
import type { ObjectiveStatus } from "../../types";

export type NewObjectiveStatus = {
  name: string;
  category: StateCategory;
  color?: string;
};

export const createObjectiveStatusAction = async (
  newStatus: NewObjectiveStatus,
) => {
  const response = await post<NewObjectiveStatus, ApiResponse<ObjectiveStatus>>(
    "objective-statuses",
    newStatus,
  );

  revalidateTag(objectiveTags.statuses());

  if (!response.data) {
    throw new Error("Failed to create objective status");
  }

  return response.data;
};
