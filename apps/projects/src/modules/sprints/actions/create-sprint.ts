"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { NewSprint, Sprint } from "../types";

export const createSprintAction = async (params: NewSprint) => {
  try {
    const sprint = await post<NewSprint, ApiResponse<Sprint>>(
      "sprints",
      params,
    );
    return sprint.data!;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to create sprint");
  }
};
