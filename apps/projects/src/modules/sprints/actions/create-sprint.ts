"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { NewSprint, Sprint } from "../types";

export const createSprintAction = async (
  params: NewSprint,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const sprint = await post<NewSprint, ApiResponse<Sprint>>(
      "sprints",
      params,
      ctx,
    );
    return sprint;
  } catch (error) {
    return getApiError(error);
  }
};
