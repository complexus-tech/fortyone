"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { NewObjective } from "../types";
import type { Objective } from "../types";

export const createObjective = async (
  params: NewObjective,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const res = await post<NewObjective, ApiResponse<Objective>>(
      "objectives",
      params,
      ctx,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
