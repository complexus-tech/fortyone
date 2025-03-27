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
    return sprint;
  } catch (error) {
    return getApiError(error);
  }
};
