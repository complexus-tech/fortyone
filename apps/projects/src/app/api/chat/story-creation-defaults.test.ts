/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { getSubscription } from "@/lib/queries/subscriptions/get-subscription";
import { getAutomationPreferences } from "@/lib/queries/users/automation-preferences";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
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
const getSubscriptionMock = jest.mocked(getSubscription);
const getWorkspaceMock = jest.mocked(getWorkspace);
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
    getSubscriptionMock.mockResolvedValue(subscription);
    getWorkspaceMock.mockResolvedValue({ trialEndsOn: null } as Awaited<
      ReturnType<typeof getWorkspace>
    >);
  });

  it("uses the signed-in user's preference when background Maya is available", async () => {
    getAutomationPreferencesMock.mockResolvedValue({
      ...preferences,
      autoScheduling: false,
    });

    await expect(resolveStoryCreationDefaults({ ctx })).resolves.toEqual({
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
    expect(getSubscriptionMock).toHaveBeenCalledWith(ctx);
  });

  it("fails safe with scheduling off for a free workspace without a trial", async () => {
    getSubscriptionMock.mockResolvedValue({
      ...subscription,
      tier: "free",
    });

    await expect(resolveStoryCreationDefaults({ ctx })).resolves.toEqual({
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
    getSubscriptionMock.mockResolvedValue(null);
    getWorkspaceMock.mockResolvedValue({
      trialEndsOn: "2099-01-01T00:00:00.000Z",
    } as Awaited<ReturnType<typeof getWorkspace>>);

    await expect(resolveStoryCreationDefaults({ ctx })).resolves.toEqual({
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

  it("fails safe with scheduling off when subscription lookup fails and no trial is active", async () => {
    getSubscriptionMock.mockRejectedValue(
      new Error("Subscription unavailable"),
    );

    getWorkspaceMock.mockResolvedValue({
      trialEndsOn: "2020-01-01T00:00:00.000Z",
    } as Awaited<ReturnType<typeof getWorkspace>>);

    await expect(resolveStoryCreationDefaults({ ctx })).resolves.toEqual({
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

  it("does not claim scheduling is enabled when preferences cannot be loaded", async () => {
    getAutomationPreferencesMock.mockRejectedValue(
      new Error("Preferences unavailable"),
    );

    await expect(resolveStoryCreationDefaults({ ctx })).resolves.toEqual({
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
