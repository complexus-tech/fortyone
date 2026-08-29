import {
  getCommentDeliveryLabel,
  shouldPollRequestThread,
} from "./delivery-status";
import type {
  IntegrationRequestComment,
  IntegrationRequestThreadActivity,
} from "./types";

const comment = (
  deliveryStatus: IntegrationRequestComment["deliveryStatus"],
): IntegrationRequestComment => ({
  id: deliveryStatus ?? "unset",
  threadId: "thread",
  direction: "outbound",
  authorName: "Joseph",
  deliveryStatus,
  body: "Update",
  createdAt: "2026-08-09T00:00:00Z",
  updatedAt: "2026-08-09T00:00:00Z",
});

const activity = (
  comments: IntegrationRequestComment[],
): IntegrationRequestThreadActivity => ({
  thread: {
    id: "thread",
    integrationRequestId: "request",
    teamId: "team",
    provider: "slack",
    externalChannelId: "channel",
    externalThreadId: "thread-ts",
    requestTitle: "Request",
    createdAt: "2026-08-09T00:00:00Z",
    updatedAt: "2026-08-09T00:00:00Z",
  },
  comments,
});

describe("integration request comment delivery state", () => {
  it.each([
    ["sending", "Sending"],
    ["retrying", "Retrying"],
    ["failed", "Failed"],
    ["not-sent", "Not sent"],
    ["sent", "Sent"],
  ] as const)("renders %s as %s", (status, label) => {
    expect(getCommentDeliveryLabel(status)).toBe(label);
  });

  it("polls only while an outbound delivery can still change", () => {
    expect(shouldPollRequestThread(activity([comment("sending")]))).toBe(true);
    expect(shouldPollRequestThread(activity([comment("retrying")]))).toBe(true);
    expect(shouldPollRequestThread(activity([comment("sent")]))).toBe(false);
    expect(shouldPollRequestThread(activity([comment("failed")]))).toBe(false);
    expect(shouldPollRequestThread(undefined)).toBe(false);
  });
});
