"use server";

import { revalidateTag } from "next/cache";
import { statusTags } from "@/constants/keys";
import { remove } from "@/lib/http";

export const deleteStateAction = async (stateId: string) => {
  const _ = await remove(`states/${stateId}`);
  revalidateTag(statusTags.lists());
  return stateId;
};
