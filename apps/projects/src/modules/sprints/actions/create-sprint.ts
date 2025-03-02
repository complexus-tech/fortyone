"use server";

import { revalidateTag } from "next/cache";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { sprintTags } from "@/constants/keys";
import { getApiError } from "@/utils";
import type { NewSprint, Sprint } from "../types";

export const createSprintAction = async (params: NewSprint) => {
  try {
    const sprint = await post<NewSprint, ApiResponse<Sprint>>(
      "sprints",
      params,
    );
    revalidateTag(sprintTags.team(params.teamId));
    revalidateTag(sprintTags.lists());
    return sprint.data!;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to create sprint");
  }
};
