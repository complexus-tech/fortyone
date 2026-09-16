import type { Subscription, Workspace } from "@/types";
import {
  TIER_LIMITS,
  type SubscriptionTier,
} from "@/shared/maya-usage/plan-limits";

export class ChatAdmissionError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message);
    this.name = "ChatAdmissionError";
  }
}

export const resolveChatTimezone = (timezone: unknown): string => {
  if (timezone === undefined) return "UTC";
  if (typeof timezone !== "string" || timezone.length > 100)
    throw new ChatAdmissionError("A valid IANA timezone is required.", 400);
  try {
    return new Intl.DateTimeFormat("en", {
      timeZone: timezone,
    }).resolvedOptions().timeZone;
  } catch {
    throw new ChatAdmissionError("A valid IANA timezone is required.", 400);
  }
};

export const resolveChatMessageLimit = (
  workspace: Pick<Workspace, "trialEndsOn">,
  subscription: Pick<Subscription, "tier" | "status"> | null,
  now = Date.now(),
) => {
  let tier: SubscriptionTier =
    workspace.trialEndsOn && Date.parse(workspace.trialEndsOn) > now
      ? "trial"
      : "free";
  if (
    subscription &&
    subscription.tier !== "free" &&
    ["active", "trialing", "past_due"].includes(subscription.status)
  )
    tier = subscription.tier;
  return TIER_LIMITS[tier].maxAiMessages;
};

export const assertChatAdmission = ({
  current,
  limit,
  isInternal,
}: {
  current: number;
  limit: number;
  isInternal: boolean;
}) => {
  if (!Number.isSafeInteger(current) || current < 0)
    throw new ChatAdmissionError("Maya usage is temporarily unavailable.", 503);
  if (!isInternal && current >= limit)
    throw new ChatAdmissionError(
      "You have reached your workspace's monthly Maya message limit.",
      429,
    );
};
