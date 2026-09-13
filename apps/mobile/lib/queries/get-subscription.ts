import { get } from "@/lib/http";
import type { ApiResponse, Subscription } from "@/types";

export const getSubscription = async (signal?: AbortSignal) => {
  const response = await get<ApiResponse<Subscription>>("subscription", {
    signal,
  });
  return response.data ?? null;
};
