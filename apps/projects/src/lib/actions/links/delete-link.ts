"use server";

import { remove } from "@/lib/http";

export const deleteLinkAction = async (linkId: string) => {
  const _ = await remove(`links/${linkId}`);
  return linkId;
};
