"use server";

import { revalidateTag } from "next/cache";
import { put } from "@/lib/http";
import type { KeyResultUpdate } from "../types";
import { objectiveTags } from "../constants";

export const updateKeyResult = async (
  keyResultId: string,
  objectiveId: string,
  params: KeyResultUpdate,
) => {
  await put<KeyResultUpdate, null>(`key-results/${keyResultId}`, params);
  revalidateTag(objectiveTags.keyResults(objectiveId));
};
