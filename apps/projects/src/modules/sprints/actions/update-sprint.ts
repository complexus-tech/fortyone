"use server";

import { revalidateTag } from "next/cache";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { sprintTags } from "@/constants/keys";
import type { Sprint, UpdateSprint } from "../types";

export const updateSprintAction = async (id: string, params: UpdateSprint) => {
  const sprint = await post<UpdateSprint, ApiResponse<Sprint>>(
    `sprints/${id}`,
    params,
  );
  revalidateTag(sprintTags.lists());

  return sprint.data!;
};
