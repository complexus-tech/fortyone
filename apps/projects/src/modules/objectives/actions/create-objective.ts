"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { NewObjective, Objective } from "../types";

export const createObjective = async (params: NewObjective) => {
  try {
    const session = await auth();
    const res = await post<NewObjective, ApiResponse<Objective>>(
      "objectives",
      params,
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
