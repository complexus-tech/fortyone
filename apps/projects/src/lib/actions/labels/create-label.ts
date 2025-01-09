"use server";

import { revalidateTag } from "next/cache";
import { post } from "@/lib/http";
import type { ApiResponse, Label } from "@/types";
import { labelTags } from "@/constants/keys";

export type NewLabel = {
  name: string;
  color: string;
  teamId?: string;
};

export const createLabelAction = async (newLabel: NewLabel) => {
  const label = await post<NewLabel, ApiResponse<Label>>("labels", newLabel);
  revalidateTag(labelTags.lists());
  return label.data!;
};
