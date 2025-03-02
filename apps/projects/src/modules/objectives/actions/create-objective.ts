"use server";

import { revalidateTag } from "next/cache";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { NewObjective, Objective } from "../types";
import { objectiveTags } from "../constants";

export const createObjective = async (params: NewObjective) => {
  try {
    const res = await post<NewObjective, ApiResponse<Objective>>(
      "objectives",
      params,
    );
    revalidateTag(objectiveTags.list());
    return res.data;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to create objective");
  }
};
