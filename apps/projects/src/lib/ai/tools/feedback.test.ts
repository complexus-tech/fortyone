/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ToolExecutionOptions, ToolExecuteFunction } from "ai";
import type { TeamFeedbackItem } from "@/modules/team-feedback/types";
import { auth } from "@/auth";
import { getTeamFeedbackItem } from "@/modules/team-feedback/queries/get-feedback";
import { getTeamFeedbackPage } from "@/modules/team-feedback/queries/get-team-feedback";
import { getCustomerFeedbackTool, listCustomerFeedbackTool } from "./feedback";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/modules/team-feedback/queries/get-feedback", () => ({
  getTeamFeedbackItem: jest.fn(),
}));

jest.mock("@/modules/team-feedback/queries/get-team-feedback", () => ({
  getTeamFeedbackPage: jest.fn(),
}));

const authMock = jest.mocked(auth);
const getTeamFeedbackItemMock = jest.mocked(getTeamFeedbackItem);
const getTeamFeedbackPageMock = jest.mocked(getTeamFeedbackPage);

const session = {
  user: {
    id: "user-1",
    name: "Joseph Mukorivo",
    email: "joseph@example.com",
    image: null,
    username: "joseph",
    fullName: "Joseph Mukorivo",
    isInternal: false,
    lastUsedWorkspaceId: "workspace-1",
  },
};

const toolOptions: ToolExecutionOptions = {
  toolCallId: "tool-call-1",
  messages: [],
  experimental_context: { workspaceSlug: "complexus" },
};

const executeTool = async <Input, Output>(
  execute: ToolExecuteFunction<Input, Output> | undefined,
  input: Input,
  options: ToolExecutionOptions = toolOptions,
): Promise<Output> => {
  if (!execute) throw new Error("Tool does not have an execute function");

  const result = execute(input, options);
  if (
    typeof result === "object" &&
    result !== null &&
    Symbol.asyncIterator in result
  ) {
    throw new Error("Streaming tool results are not supported by this test");
  }

  return (await result) as Output;
};

const feedback: TeamFeedbackItem = {
  id: "feedback-1",
  workspaceId: "workspace-1",
  portalId: "portal-1",
  boardId: "board-1",
  authorId: "customer-1",
  authorName: "Ada Ndlovu",
  authorAvatar: null,
  title: "Publish a customer roadmap",
  description:
    "Let customers see which requests are planned, in progress, and completed.",
  slug: "publish-a-customer-roadmap",
  status: "planned",
  voteCount: 12,
  upvoteCount: 14,
  downvoteCount: 2,
  commentCount: 1,
  readAt: "2026-07-20T08:30:00.000Z",
  roadmapSummary: "A public roadmap is planned for the next release.",
  createdAt: "2026-07-18T08:30:00.000Z",
  updatedAt: "2026-07-20T08:30:00.000Z",
  board: {
    id: "board-1",
    workspaceId: "workspace-1",
    portalId: "portal-1",
    teamId: "team-1",
    name: "Product requests",
    slug: "product-requests",
    color: "#16A34A",
    orderIndex: 0,
    createdAt: "2026-07-01T08:30:00.000Z",
    updatedAt: "2026-07-01T08:30:00.000Z",
  },
  comments: [
    {
      id: "comment-1",
      workspaceId: "workspace-1",
      itemId: "feedback-1",
      authorId: "customer-2",
      authorName: "Tariro Moyo",
      authorAvatar: null,
      body: "This would help us understand what is coming next.",
      createdAt: "2026-07-19T08:30:00.000Z",
      updatedAt: "2026-07-19T08:30:00.000Z",
    },
  ],
  storyLinks: [
    {
      id: "link-1",
      workspaceId: "workspace-1",
      itemId: "feedback-1",
      storyId: "story-1",
      storyTitle: "Build the public roadmap",
      relationship: "created_from",
      isPrimary: true,
      createdByUserId: "user-1",
      createdAt: "2026-07-20T08:30:00.000Z",
    },
  ],
};

describe("Maya feedback tools", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
  });

  it("lists scoped feedback for each requested team with readable summaries", async () => {
    getTeamFeedbackPageMock.mockResolvedValue({
      feedback: [feedback],
      pagination: {
        page: 2,
        pageSize: 10,
        hasMore: false,
        nextPage: 3,
      },
    });

    const result = await executeTool(listCustomerFeedbackTool.execute, {
      teamIds: ["team-1", "team-1"],
      status: "planned" as const,
      search: "  roadmap  ",
      page: 2,
      pageSize: 10,
    });

    expect(getTeamFeedbackPageMock).toHaveBeenCalledTimes(1);
    expect(getTeamFeedbackPageMock).toHaveBeenCalledWith(
      "team-1",
      { session, workspaceSlug: "complexus" },
      "planned",
      "roadmap",
      2,
      10,
    );
    expect(result).toMatchObject({
      success: true,
      count: 1,
      filters: { search: "roadmap", status: "planned" },
      teams: [
        {
          teamId: "team-1",
          pagination: { hasMore: false, nextPage: null },
          feedback: [
            {
              id: "feedback-1",
              title: "Publish a customer roadmap",
              statusLabel: "Planned",
              author: "Ada Ndlovu",
              board: { name: "Product requests" },
              voteScore: 12,
              isRead: true,
              primaryStory: {
                id: "story-1",
                title: "Build the public roadmap",
                relationship: "created_from",
              },
            },
          ],
        },
      ],
    });
    expect(result).not.toHaveProperty("teams.0.feedback.0.comments");
    expect(result).not.toHaveProperty("teams.0.feedback.0.workspaceId");
    expect(result).not.toHaveProperty("teams.0.feedback.0.portalId");
    expect(result).not.toHaveProperty("teams.0.feedback.0.authorAvatar");
    expect(result).not.toHaveProperty("teams.0.feedback.0.board.color");
    expect(result).not.toHaveProperty("teams.0.feedback.0.board.slug");
  });

  it("uses the active queue and safe pagination defaults", async () => {
    getTeamFeedbackPageMock.mockResolvedValue({
      feedback: [],
      pagination: {
        page: 1,
        pageSize: 20,
        hasMore: true,
        nextPage: 2,
      },
    });

    const result = await executeTool(listCustomerFeedbackTool.execute, {
      teamIds: ["team-1"],
    });

    expect(getTeamFeedbackPageMock).toHaveBeenCalledWith(
      "team-1",
      { session, workspaceSlug: "complexus" },
      "active",
      "",
      1,
      20,
    );
    expect(result).toMatchObject({
      success: true,
      count: 0,
      hasMore: true,
      message:
        "Returned 0 feedback items. More feedback is available on the next page.",
    });
  });

  it("returns full feedback discussion and linked work for one item", async () => {
    getTeamFeedbackItemMock.mockResolvedValue(feedback);

    const result = await executeTool(getCustomerFeedbackTool.execute, {
      feedbackId: "feedback-1",
    });

    expect(getTeamFeedbackItemMock).toHaveBeenCalledWith("feedback-1", {
      session,
      workspaceSlug: "complexus",
    });
    expect(result).toMatchObject({
      success: true,
      feedback: {
        id: "feedback-1",
        teamId: "team-1",
        description: feedback.description,
        roadmapSummary: feedback.roadmapSummary,
        comments: [
          {
            author: "Tariro Moyo",
            body: "This would help us understand what is coming next.",
          },
        ],
      },
    });
  });

  it("bounds untrusted detail text and returns only the newest comments", async () => {
    const comments = Array.from({ length: 30 }, (_, index) => ({
      ...feedback.comments[0],
      id: `comment-${index + 1}`,
      body: "x".repeat(2200),
      createdAt: `2026-07-${String(index + 1).padStart(2, "0")}T08:30:00.000Z`,
    }));
    getTeamFeedbackItemMock.mockResolvedValue({
      ...feedback,
      description: "d".repeat(5200),
      roadmapSummary: "r".repeat(2200),
      commentCount: comments.length,
      comments,
    });

    const result = await executeTool(getCustomerFeedbackTool.execute, {
      feedbackId: "feedback-1",
    });

    expect(result).toMatchObject({
      success: true,
      feedback: {
        descriptionTruncated: true,
        roadmapSummaryTruncated: true,
        commentsReturned: 25,
        commentsOmitted: 5,
        linkedStories: [
          {
            id: "story-1",
            title: "Build the public roadmap",
          },
        ],
      },
    });
    expect(result).toHaveProperty(
      "feedback.description",
      `${"d".repeat(4997)}...`,
    );
    expect(result).toHaveProperty(
      "feedback.roadmapSummary",
      `${"r".repeat(1997)}...`,
    );
    expect(result).toHaveProperty(
      "feedback.comments.0.body",
      `${"x".repeat(1997)}...`,
    );
    expect(result).toHaveProperty("feedback.comments.0.bodyTruncated", true);
    expect(result).toHaveProperty(
      "feedback.comments.0.createdAt",
      comments[5]?.createdAt,
    );
    expect(result).not.toHaveProperty("feedback.comments.0.id");
  });

  it("rejects feedback access without a session", async () => {
    authMock.mockResolvedValue(null);

    const result = await executeTool(listCustomerFeedbackTool.execute, {
      teamIds: ["team-1"],
    });

    expect(result).toEqual({
      success: false,
      error: "Authentication required to access customer feedback",
    });
    expect(getTeamFeedbackPageMock).not.toHaveBeenCalled();
  });

  it("rejects feedback access without a workspace context", async () => {
    const result = await executeTool(
      listCustomerFeedbackTool.execute,
      { teamIds: ["team-1"] },
      {
        ...toolOptions,
        experimental_context: undefined,
      },
    );

    expect(result).toEqual({
      success: false,
      error: "Workspace context is required",
    });
    expect(getTeamFeedbackPageMock).not.toHaveBeenCalled();
  });

  it("preserves API permission failures", async () => {
    getTeamFeedbackItemMock.mockRejectedValue(
      new Error("You do not have access to this team"),
    );

    const result = await executeTool(getCustomerFeedbackTool.execute, {
      feedbackId: "feedback-1",
    });

    expect(result).toEqual({
      success: false,
      error: "You do not have access to this team",
    });
  });
});
