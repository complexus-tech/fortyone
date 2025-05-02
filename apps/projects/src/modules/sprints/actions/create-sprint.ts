"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { NewSprint, Sprint } from "../types";

export const createSprintAction = async (params: NewSprint) => {
  try {
    const session = await auth();
    const sprint = await post<NewSprint, ApiResponse<Sprint>>(
      "sprints",
      params,
      session!,
    );
    return sprint;
  } catch (error) {
    return getApiError(error);
  }
};
