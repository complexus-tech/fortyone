import { get } from "@/lib/http";
import type { ApiResponse, Subscription } from "@/types";

export const getSubscription = async () => {
  try {
    const subscription = await get<ApiResponse<Subscription>>("subscription");
    return subscription.data || null;
  } catch {
    return null;
  }
};
