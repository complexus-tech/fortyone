/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { auth } from "@/auth";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { acceptAllIntegrationRequestsAction } from "@/modules/integration-requests/actions/accept-all";
import { declineAllIntegrationRequestsAction } from "@/modules/integration-requests/actions/decline-all";
import { bulkUpdateAction } from "@/modules/stories/actions/bulk-update-stories";
import {
  acceptAllIntegrationRequestsTool,
  declineAllIntegrationRequestsTool,
  listIntegrationRequestsTool,
  updateIntegrationRequestTool,
} from "./integration-requests";
import { bulkUpdateStories } from "./stories/bulk-update-stories";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/queries/workspaces/get-workspace", () => ({
  getWorkspace: jest.fn(),
}));
jest.mock("@/modules/stories/actions/bulk-update-stories", () => ({
  bulkUpdateAction: jest.fn(),
}));
jest.mock("@/modules/integration-requests/actions/accept-all", () => ({
  acceptAllIntegrationRequestsAction: jest.fn(),
}));
jest.mock("@/modules/integration-requests/actions/accept", () => ({
  acceptIntegrationRequestAction: jest.fn(),
}));
jest.mock("@/modules/integration-requests/actions/decline-all", () => ({
  declineAllIntegrationRequestsAction: jest.fn(),
}));
jest.mock("@/modules/integration-requests/actions/decline", () => ({
  declineIntegrationRequestAction: jest.fn(),
}));
jest.mock("@/modules/integration-requests/actions/post-github-comment", () => ({
  postRequestGitHubCommentAction: jest.fn(),
}));
jest.mock("@/modules/integration-requests/actions/update", () => ({
  updateIntegrationRequestAction: jest.fn(),
}));
jest.mock("@/modules/integration-requests/queries/get-request", () => ({
  getIntegrationRequest: jest.fn(),
}));
jest.mock(
  "@/modules/integration-requests/queries/get-request-github-comments",
  () => ({ getRequestGitHubComments: jest.fn() }),
);
jest.mock("@/modules/integration-requests/queries/get-team-requests", () => ({
  getTeamIntegrationRequestsPage: jest.fn(),
}));

const execute = async (
  definition: { execute?: (input: never, options: never) => unknown },
  input: Record<string, unknown>,
) => {
  if (!definition.execute) throw new Error("Tool has no execute function.");

  return definition.execute(
    input as never,
    {
      experimental_context: { workspaceSlug: "acme" },
      messages: [],
      toolCallId: "bulk-call",
    } as never,
  ) as Promise<Record<string, unknown>>;
};

const parseToolInput = (definition: unknown, input: Record<string, unknown>) =>
  (
    definition as {
      inputSchema: { safeParse: (value: unknown) => { success: boolean } };
    }
  ).inputSchema.safeParse(input);

describe("bulk mutation result contracts", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.mocked(auth).mockResolvedValue({ user: { id: "user-1" } } as never);
    jest.mocked(getWorkspace).mockResolvedValue({ userRole: "admin" } as never);
  });

  it("bounds integration request team fan-out and validates UUIDs", () => {
    const teamId = "18227a68-8e35-4aad-a8ee-fb6b1ef9feee";

    expect(
      parseToolInput(listIntegrationRequestsTool, { teamIds: [teamId] })
        .success,
    ).toBe(true);
    expect(
      parseToolInput(listIntegrationRequestsTool, {
        teamIds: ["not-a-uuid"],
      }).success,
    ).toBe(false);
    expect(
      parseToolInput(listIntegrationRequestsTool, {
        teamIds: Array.from(
          { length: 21 },
          (_, index) =>
            `18227a68-8e35-4aad-a8ee-${String(index).padStart(12, "0")}`,
        ),
      }).success,
    ).toBe(false);
    expect(
      parseToolInput(listIntegrationRequestsTool, {
        teamIds: [teamId, teamId],
      }).success,
    ).toBe(false);
  });

  it("rejects no-op integration request updates but preserves null clears", () => {
    expect(
      parseToolInput(updateIntegrationRequestTool, {
        requestId: "request-1",
        confirmed: true,
      }).success,
    ).toBe(false);
    expect(
      parseToolInput(updateIntegrationRequestTool, {
        requestId: "request-1",
        confirmed: true,
        estimateValue: null,
      }).success,
    ).toBe(true);
  });

  it("returns every story update outcome and marks partial completion", async () => {
    jest.mocked(bulkUpdateAction).mockResolvedValue({
      data: {
        totalCount: 3,
        succeededCount: 2,
        failedCount: 1,
        partial: true,
        items: [
          { storyId: "story-1", success: true },
          { storyId: "story-2", success: false, error: "Story not found" },
          { storyId: "story-3", success: true },
        ],
      },
    } as never);

    const result = await execute(bulkUpdateStories, {
      storyIds: ["story-1", "story-2", "story-3"],
      confirmed: true,
      updateData: { priority: "High" },
    });

    expect(result).toMatchObject({
      success: false,
      partial: true,
      result: {
        totalCount: 3,
        succeededCount: 2,
        failedCount: 1,
        items: [
          { storyId: "story-1", success: true },
          { storyId: "story-2", success: false, error: "Story not found" },
          { storyId: "story-3", success: true },
        ],
      },
    });
    expect(result.message).toBe("Updated 2 of 3 stories; 1 failed.");
  });

  it.each([
    [
      "accept",
      acceptAllIntegrationRequestsTool,
      acceptAllIntegrationRequestsAction,
    ],
    [
      "decline",
      declineAllIntegrationRequestsTool,
      declineAllIntegrationRequestsAction,
    ],
  ] as const)(
    "returns per-request outcomes for partial bulk %s",
    async (operation, toolDefinition, action) => {
      jest.mocked(action).mockResolvedValue({
        data: {
          count: 1,
          requestIds: ["request-1"],
          totalCount: 2,
          succeededCount: 1,
          failedCount: 1,
          partial: true,
          items: [
            {
              requestId: "request-1",
              success: true,
              status: operation === "accept" ? "accepted" : "declined",
            },
            {
              requestId: "request-2",
              success: false,
              status: "failed",
              error: "Request is no longer pending",
            },
          ],
        },
      } as never);

      const result = await execute(toolDefinition, {
        teamId: "team-1",
        confirmed: true,
      });

      expect(result).toMatchObject({
        success: false,
        partial: true,
        result: {
          totalCount: 2,
          succeededCount: 1,
          failedCount: 1,
          items: [
            { requestId: "request-1", success: true },
            {
              requestId: "request-2",
              success: false,
              error: "Request is no longer pending",
            },
          ],
        },
      });
      expect(result.message).toContain("1 of 2");
      expect(result.message).toContain("1 failed");
    },
  );
});
