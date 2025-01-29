"use server";

import { revalidateTag } from "next/cache";
import { remove } from "@/lib/http";
import { objectiveTags } from "../constants";

export const deleteObjective = async (objectiveId: string) => {
  await remove(`objectives/${objectiveId}`);
  revalidateTag(objectiveTags.list());
  revalidateTag(objectiveTags.objective(objectiveId));
};
