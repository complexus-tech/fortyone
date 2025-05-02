import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse, Subscription } from "@/types";

export const getSubscription = async (session: Session) => {
  try {
    const subscription = await get<ApiResponse<Subscription>>(
      "subscription",
      session,
    );
    return subscription.data!;
  } catch {
    return null;
  }
};
