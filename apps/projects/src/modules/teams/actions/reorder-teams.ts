"use server";
import { patch } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { ReorderTeamsInput } from "../types";

export async function reorderTeamsAction(input: ReorderTeamsInput) {
  try {
    const session = await auth();
    const response = await patch<ReorderTeamsInput, ApiResponse<void>>(
      "teams/order",
      input,
      session!,
    );
    return response;
  } catch (error) {
    return getApiError(error);
  }
}
