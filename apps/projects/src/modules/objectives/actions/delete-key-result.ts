"use server";

import { revalidateTag } from "next/cache";
import { remove } from "@/lib/http";
import { objectiveTags } from "../constants";

export const deleteKeyResult = async (
  keyResultId: string,
  objectiveId: string,
) => {
  await remove(`key-results/${keyResultId}`);
  revalidateTag(objectiveTags.keyResults(objectiveId));
};
