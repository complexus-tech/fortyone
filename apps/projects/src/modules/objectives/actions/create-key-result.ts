"use server";

import { revalidateTag } from "next/cache";
import { post } from "@/lib/http";
import type { NewObjectiveKeyResult } from "../types";
import { objectiveTags } from "../constants";

export const createKeyResult = async (params: NewObjectiveKeyResult) => {
  await post<NewObjectiveKeyResult, null>("key-results", params);
  revalidateTag(objectiveTags.keyResults(params.objectiveId));
};
