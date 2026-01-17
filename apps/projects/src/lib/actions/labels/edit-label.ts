"use server";

import { put } from "@/lib/http";
import type { ApiResponse, Label } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export type UpdateLabel = {
  name: string;
  color: string;
};

export const editLabelAction = async (
  labelId: string,
  updates: UpdateLabel,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const label = await put<UpdateLabel, ApiResponse<Label>>(
      `labels/${labelId}`,
      updates,
      ctx,
    );
    return label;
  } catch (error) {
    return getApiError(error);
  }
};
