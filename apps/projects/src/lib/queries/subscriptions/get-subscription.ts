import { get } from "@/lib/http";
import type { ApiResponse, Subscription } from "@/types";

export const getSubscription = async () => {
  try {
    const subscription = await get<ApiResponse<Subscription>>("subscription");
    return subscription.data!;
  } catch (error) {
    const freeSubscription: Subscription = {
      workspaceId: "",
      stripeCustomerId: "",
      stripeSubscriptionId: "",
      status: "active",
      tier: "free",
      seatCount: 1,
      createdAt: "",
      updatedAt: "",
    };
    return freeSubscription;
  }
};
