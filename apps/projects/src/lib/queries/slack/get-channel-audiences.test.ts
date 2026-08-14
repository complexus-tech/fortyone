/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { SlackChannelAudience } from "@/modules/settings/workspace/integrations/slack/types";
import { get } from "@/lib/http";
import { getSlackChannelAudiences } from "./get-channel-audiences";

jest.mock("@/lib/http", () => ({
  get: jest.fn(),
}));

const mockGet = get as unknown as {
  mockResolvedValue: (value: unknown) => void;
};

const audience: SlackChannelAudience = {
  channel: {
    id: "channel-record-1",
    slackChannelId: "C123",
    name: "product",
    isPrivate: false,
    isArchived: false,
    isMember: true,
    isActive: true,
    createdAt: "2026-08-14T09:00:00Z",
    updatedAt: "2026-08-14T09:00:00Z",
  },
  isConfigured: true,
  teamIds: [],
};

describe("getSlackChannelAudiences", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("returns audiences with explicit configuration state", async () => {
    mockGet.mockResolvedValue({ data: [audience] });

    await expect(
      getSlackChannelAudiences({ workspaceSlug: "complexus" }),
    ).resolves.toEqual([audience]);
  });

  it("fails closed when an older API omits configuration state", async () => {
    const { isConfigured: _isConfigured, ...legacyAudience } = audience;
    mockGet.mockResolvedValue({ data: [legacyAudience] });

    await expect(
      getSlackChannelAudiences({ workspaceSlug: "complexus" }),
    ).rejects.toThrow("Slack channel configuration is not available");
  });
});
