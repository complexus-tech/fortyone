"use server";
import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { ReorderTeamsInput } from "../types";

export async function reorderTeamsAction(
  input: ReorderTeamsInput,
  workspaceSlug: string,
) {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const response = await put<ReorderTeamsInput, ApiResponse<void>>(
      "teams/order",
      input,
      ctx,
    );
    return response;
  } catch (error) {
    return getApiError(error);
  }
}
