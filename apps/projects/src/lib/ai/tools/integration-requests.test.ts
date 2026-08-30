/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { auth } from "@/auth";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { acceptAllIntegrationRequestsAction } from "@/modules/integration-requests/actions/accept-all";
import { acceptIntegrationRequestAction } from "@/modules/integration-requests/actions/accept";
import { declineAllIntegrationRequestsAction } from "@/modules/integration-requests/actions/decline-all";
import { declineIntegrationRequestAction } from "@/modules/integration-requests/actions/decline";
import { postRequestGitHubCommentAction } from "@/modules/integration-requests/actions/post-github-comment";
import { updateIntegrationRequestAction } from "@/modules/integration-requests/actions/update";
import { getIntegrationRequest } from "@/modules/integration-requests/queries/get-request";
import { getRequestGitHubComments } from "@/modules/integration-requests/queries/get-request-github-comments";
import { getTeamIntegrationRequestsPage } from "@/modules/integration-requests/queries/get-team-requests";
import {
  acceptAllIntegrationRequestsTool,
  acceptIntegrationRequestTool,
  declineAllIntegrationRequestsTool,
  declineIntegrationRequestTool,
  getIntegrationRequestTool,
  listIntegrationRequestsTool,
  postRequestGitHubCommentTool,
  updateIntegrationRequestTool,
} from "./integration-requests";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/queries/workspaces/get-workspace", () => ({
  getWorkspace: jest.fn(),
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

const execute = async (definition: unknown, input: Record<string, unknown>) => {
  const executeTool = (
    definition as {
      execute?: (toolInput: never, options: never) => unknown;
    }
  ).execute;

  if (!executeTool) throw new Error("Tool has no execute function.");

  return executeTool(
    input as never,
    {
      experimental_context: { workspaceSlug: "acme" },
      messages: [],
      toolCallId: "integration-request-call",
    } as never,
  ) as Promise<Record<string, unknown>>;
};

const integrationRequest = {
  createdAt: "2026-08-30T00:00:00.000Z",
  id: "request-1",
  priority: "High",
  provider: "github",
  sourceType: "issue",
  status: "pending",
  teamId: "team-1",
  title: "Improve onboarding",
  updatedAt: "2026-08-30T00:00:00.000Z",
};

describe("integration request tools", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.mocked(auth).mockResolvedValue({ user: { id: "user-1" } } as never);
    jest.mocked(getWorkspace).mockResolvedValue({ userRole: "admin" } as never);
  });

  it("short-circuits every mutation before authentication or side effects", async () => {
    const confirmationCases = [
      {
        action: updateIntegrationRequestAction,
        input: { requestId: "request-1", title: "Update title" },
        message: "Please confirm before I update this integration request.",
        toolDefinition: updateIntegrationRequestTool,
      },
      {
        action: acceptIntegrationRequestAction,
        input: { requestId: "request-1" },
        message: "Please confirm before I accept this integration request.",
        toolDefinition: acceptIntegrationRequestTool,
      },
      {
        action: declineIntegrationRequestAction,
        input: { requestId: "request-1" },
        message: "Please confirm before I decline this integration request.",
        toolDefinition: declineIntegrationRequestTool,
      },
      {
        action: acceptAllIntegrationRequestsAction,
        input: { teamId: "team-1" },
        message:
          "Please confirm before I accept all pending integration requests in this team.",
        toolDefinition: acceptAllIntegrationRequestsTool,
      },
      {
        action: declineAllIntegrationRequestsAction,
        input: { teamId: "team-1" },
        message:
          "Please confirm before I decline all pending integration requests in this team.",
        toolDefinition: declineAllIntegrationRequestsTool,
      },
      {
        action: postRequestGitHubCommentAction,
        input: { requestId: "request-1", body: "Please investigate." },
        message:
          "Please confirm before I post this comment to the request's GitHub issue.",
        toolDefinition: postRequestGitHubCommentTool,
      },
    ];

    await Promise.all(
      confirmationCases.map(
        async ({ action, input, message, toolDefinition }) => {
          await expect(execute(toolDefinition, input)).resolves.toEqual({
            success: false,
            needsConfirmation: true,
            message,
          });
          expect(action).not.toHaveBeenCalled();
        },
      ),
    );

    expect(auth).not.toHaveBeenCalled();
    expect(getWorkspace).not.toHaveBeenCalled();
  });

  it.each([
    [
      "accept",
      acceptIntegrationRequestTool,
      { requestId: "request-1", confirmed: true },
      acceptIntegrationRequestAction,
      "Guests cannot accept requests",
    ],
    [
      "decline",
      declineIntegrationRequestTool,
      { requestId: "request-1", confirmed: true },
      declineIntegrationRequestAction,
      "Guests cannot decline requests",
    ],
    [
      "bulk accept",
      acceptAllIntegrationRequestsTool,
      { teamId: "team-1", confirmed: true },
      acceptAllIntegrationRequestsAction,
      "Guests cannot accept requests",
    ],
    [
      "bulk decline",
      declineAllIntegrationRequestsTool,
      { teamId: "team-1", confirmed: true },
      declineAllIntegrationRequestsAction,
      "Guests cannot decline requests",
    ],
  ])(
    "does not let a guest %s integration requests",
    async (_operation, toolDefinition, input, action, error) => {
      jest
        .mocked(getWorkspace)
        .mockResolvedValue({ userRole: "guest" } as never);

      await expect(execute(toolDefinition, input)).resolves.toEqual({
        success: false,
        error,
      });
      expect(action).not.toHaveBeenCalled();
    },
  );

  it("preserves null clears and time-patch normalization for confirmed updates", async () => {
    jest.mocked(updateIntegrationRequestAction).mockResolvedValue({
      data: integrationRequest,
    } as never);

    await expect(
      execute(updateIntegrationRequestTool, {
        requestId: "request-1",
        confirmed: true,
        title: "  Improve onboarding  ",
        estimatedDurationMinutes: null,
        minimumFocusBlockMinutes: 60,
      }),
    ).resolves.toMatchObject({
      success: true,
      request: integrationRequest,
    });

    expect(updateIntegrationRequestAction).toHaveBeenCalledWith(
      "request-1",
      expect.objectContaining({
        estimatedDurationMinutes: null,
        minimumFocusBlockMinutes: null,
        title: "Improve onboarding",
      }),
      "acme",
    );
  });

  it("loads optional comments only for GitHub requests", async () => {
    jest.mocked(getIntegrationRequest).mockResolvedValue({
      ...integrationRequest,
      provider: "slack",
    } as never);

    await expect(
      execute(getIntegrationRequestTool, {
        requestId: "request-1",
        includeGitHubComments: true,
      }),
    ).resolves.toMatchObject({ success: true });
    expect(getRequestGitHubComments).not.toHaveBeenCalled();

    jest
      .mocked(getIntegrationRequest)
      .mockResolvedValue(integrationRequest as never);
    jest
      .mocked(getRequestGitHubComments)
      .mockResolvedValue([{ id: 1, body: "Looks good" }] as never);

    await expect(
      execute(getIntegrationRequestTool, {
        requestId: "request-1",
        includeGitHubComments: true,
      }),
    ).resolves.toMatchObject({
      success: true,
      githubComments: [{ id: 1, body: "Looks good" }],
    });
    expect(getRequestGitHubComments).toHaveBeenCalledWith(
      "request-1",
      expect.objectContaining({ workspaceSlug: "acme" }),
    );
  });

  it("retains per-team pagination and normalized list filters", async () => {
    const teamIds = [
      "18227a68-8e35-4aad-a8ee-fb6b1ef9feee",
      "6cda5bcf-a1e4-4f6d-9c4f-e909a563afb8",
    ];
    jest.mocked(getTeamIntegrationRequestsPage).mockResolvedValue({
      pagination: {
        page: 2,
        pageSize: 50,
        totalCount: 1,
        hasMore: true,
        nextPage: 3,
      },
      requests: [integrationRequest],
    } as never);

    await expect(
      execute(listIntegrationRequestsTool, {
        teamIds,
        status: "accepted",
        provider: "github",
        assigneeId: "   ",
        createdAfter: "2026-08-01T00:00:00.000Z",
        page: 2,
        pageSize: 50,
      }),
    ).resolves.toMatchObject({
      success: true,
      count: 2,
      filters: {
        status: "accepted",
        provider: "github",
        assigneeId: "   ",
      },
      teams: [
        { teamId: teamIds[0], requests: [{ id: "request-1" }] },
        { teamId: teamIds[1], requests: [{ id: "request-1" }] },
      ],
    });

    expect(getTeamIntegrationRequestsPage).toHaveBeenNthCalledWith(
      1,
      teamIds[0],
      expect.objectContaining({ workspaceSlug: "acme" }),
      "accepted",
      2,
      50,
      {
        provider: "github",
        priority: undefined,
        assigneeId: undefined,
        createdAfter: "2026-08-01T00:00:00.000Z",
        createdBefore: undefined,
      },
    );
  });
});
