"use server";

import { revalidateTag } from "next/cache";
import { put } from "@/lib/http";
import { sprintTags } from "@/constants/keys";
import type { UpdateSprint } from "../types";

export const updateSprintAction = async (
  sprintId: string,
  updates: UpdateSprint,
) => {
  await put(`sprints/${sprintId}`, updates);
  revalidateTag(sprintTags.lists());
};
