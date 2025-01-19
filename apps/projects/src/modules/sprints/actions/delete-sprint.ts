"use server";

import { revalidateTag } from "next/cache";
import { remove } from "@/lib/http";
import { sprintTags } from "@/constants/keys";

export const deleteSprintAction = async (sprintId: string) => {
  await remove(`sprints/${sprintId}`);
  revalidateTag(sprintTags.lists());
};
