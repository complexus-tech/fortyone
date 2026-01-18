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
