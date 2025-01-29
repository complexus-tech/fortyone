"use server";

import { revalidateTag } from "next/cache";
import { put } from "@/lib/http";
import type { ObjectiveUpdate } from "../types";
import { objectiveTags } from "../constants";

export const updateObjective = async (
  objectiveId: string,
  params: ObjectiveUpdate,
) => {
  await put<ObjectiveUpdate, null>(`objectives/${objectiveId}`, params);
  revalidateTag(objectiveTags.list());
  revalidateTag(objectiveTags.objective(objectiveId));
};
