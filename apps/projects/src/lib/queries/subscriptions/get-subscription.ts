import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse, Subscription } from "@/types";

export const getSubscription = async (ctx: WorkspaceCtx) => {
  try {
    const subscription = await get<ApiResponse<Subscription>>(
      "subscription",
      ctx,
    );
    return subscription.data!;
  } catch {
    return null;
  }
};

// Unlike entitlement reads, surface failures so billing never substitutes a
// public catalogue rate for the customer's actual subscription price.
export const getSubscriptionWithPrice = async (ctx: WorkspaceCtx) => {
  const response = await get<ApiResponse<Subscription>>(
    "subscription?includePrice=true",
    ctx,
  );
  return response.data;
};
