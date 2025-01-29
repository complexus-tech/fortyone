"use server";

import { revalidateTag } from "next/cache";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { NewObjective, Objective } from "../types";
import { objectiveTags } from "../constants";

export const createObjective = async (params: NewObjective) => {
  const res = await post<NewObjective, ApiResponse<Objective>>(
    "objectives",
    params,
  );
  revalidateTag(objectiveTags.list());
  return res.data;
};
