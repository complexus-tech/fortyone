/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { getAutomationPreferences } from "@/lib/queries/users/automation-preferences";
import { resolveStoryCreationDefaults } from "./story-creation-defaults";

jest.mock("@/lib/queries/subscriptions/get-subscription", () => ({
  getSubscription: jest.fn(),
}));

jest.mock("@/lib/queries/users/automation-preferences", () => ({
  getAutomationPreferences: jest.fn(),
}));

jest.mock("@/lib/queries/workspaces/get-workspace", () => ({
  getWorkspace: jest.fn(),
}));

const getAutomationPreferencesMock = jest.mocked(getAutomationPreferences);
const ctx = {
  session: { token: "test-token" },
  workspaceSlug: "complexus",
};

const subscription = {
  workspaceId: "workspace-1",
  stripeCustomerId: "customer-1",
  stripeSubscriptionId: "subscription-1",
  status: "active" as const,
  tier: "pro" as const,
  seatCount: 5,
  billingInterval: "month" as const,
  billingEndsAt: "2026-09-27T00:00:00.000Z",
  createdAt: "2026-08-01T00:00:00.000Z",
  updatedAt: "2026-08-01T00:00:00.000Z",
};

const preferences = {
  id: "preference-1",
  autoAssignSelf: true,
  autoScheduling: true,
  assignSelfOnBranchCopy: false,
  moveStoryToStartedOnBranch: true,
  openStoryInDialog: true,
};

describe("resolveStoryCreationDefaults", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    getAutomationPreferencesMock.mockResolvedValue(preferences);
  });

  it("uses the signed-in user's preference when background Maya is available", async () => {
    getAutomationPreferencesMock.mockResolvedValue({
      ...preferences,
      autoScheduling: false,
    });

    await expect(
      resolveStoryCreationDefaults({
        ctx,
        workspace: { trialEndsOn: null },
        subscription,
      }),
    ).resolves.toEqual({
      autoSchedulingAvailable: true,
      bulkStories: {
        autoSchedulingEnabled: false,
        estimatedDurationMinutes: null,
      },
      singleStory: {
        autoSchedulingEnabled: false,
        estimatedDurationMinutes: 60,
      },
    });
    expect(getAutomationPreferencesMock).toHaveBeenCalledWith(ctx);
  });

  it("fails safe with scheduling off for a free workspace without a trial", async () => {
    await expect(
      resolveStoryCreationDefaults({
        ctx,
        workspace: { trialEndsOn: null },
        subscription: { ...subscription, tier: "free" },
      }),
    ).resolves.toEqual({
      autoSchedulingAvailable: false,
      bulkStories: {
        autoSchedulingEnabled: false,
        estimatedDurationMinutes: null,
      },
      singleStory: {
        autoSchedulingEnabled: false,
        estimatedDurationMinutes: 60,
      },
    });
  });

  it("allows the preference during an active workspace trial", async () => {
    await expect(
      resolveStoryCreationDefaults({
        ctx,
        workspace: { trialEndsOn: "2099-01-01T00:00:00.000Z" },
        subscription: null,
      }),
    ).resolves.toEqual({
      autoSchedulingAvailable: true,
      bulkStories: {
        autoSchedulingEnabled: false,
        estimatedDurationMinutes: null,
      },
      singleStory: {
        autoSchedulingEnabled: true,
        estimatedDurationMinutes: 60,
      },
    });
  });

  it("does not reload workspace or billing after trusted hydration", async () => {
    await resolveStoryCreationDefaults({
      ctx,
      workspace: { trialEndsOn: null },
      subscription: null,
    });
    expect(
      jest.requireMock("@/lib/queries/subscriptions/get-subscription")
        .getSubscription,
    ).not.toHaveBeenCalled();
    expect(
      jest.requireMock("@/lib/queries/workspaces/get-workspace").getWorkspace,
    ).not.toHaveBeenCalled();
  });

  it("does not claim scheduling is enabled when preferences cannot be loaded", async () => {
    getAutomationPreferencesMock.mockRejectedValue(
      new Error("Preferences unavailable"),
    );

    await expect(
      resolveStoryCreationDefaults({
        ctx,
        workspace: { trialEndsOn: null },
        subscription,
      }),
    ).resolves.toEqual({
      autoSchedulingAvailable: true,
      bulkStories: {
        autoSchedulingEnabled: false,
        estimatedDurationMinutes: null,
      },
      singleStory: {
        autoSchedulingEnabled: false,
        estimatedDurationMinutes: 60,
      },
    });
  });
});
