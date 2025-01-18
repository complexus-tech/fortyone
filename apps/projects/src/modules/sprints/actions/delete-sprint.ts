"use server";

import { revalidateTag } from "next/cache";
import { remove } from "@/lib/http";
import { sprintTags } from "@/constants/keys";

export const deleteSprintAction = async (id: string) => {
  await remove(`sprints/${id}`);
  revalidateTag(sprintTags.lists());
};
