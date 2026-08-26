/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { UIMessage } from "ai";
import { selectActiveTools } from "./active-tools";

const userMessage = (text: string, id = "user-message") =>
  ({
    id,
    parts: [{ text, type: "text" }],
    role: "user",
  }) as UIMessage;

describe("selectActiveTools", () => {
  it("uses a bounded story-creation tool set for bulk story requests", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/teams/product/stories",
      messages: [
        userMessage(
          "Bulk create five stories for the Product team in the current sprint.",
        ),
      ],
    });

    expect(activeTools).toEqual(
      expect.arrayContaining([
        "bulkCreateStories",
        "statuses",
        "listTeams",
        "listSprints",
      ]),
    );
    expect(activeTools).not.toContain("workspacePerformanceReportTool");
    expect(activeTools).not.toContain("getGitHubIntegrationTool");
    expect(activeTools.length).toBeLessThanOrEqual(20);
  });

  it("keeps a pending bulk-create tool available for a short confirmation", () => {
    const pendingToolMessage = {
      id: "assistant-message",
      parts: [
        {
          input: { storiesData: [] },
          output: {
            message: "Please confirm before I create these stories.",
            needsConfirmation: true,
            success: false,
          },
          state: "output-available",
          type: "tool-bulkCreateStories",
        },
      ],
      role: "assistant",
    } as unknown as UIMessage;

    const activeTools = selectActiveTools({
      currentPath: "/acme/inbox",
      messages: [pendingToolMessage, userMessage("Approved.")],
    });

    expect(activeTools).toContain("bulkCreateStories");
    expect(activeTools).not.toContain("bulkDeleteStories");
  });

  it("does not send unrelated story mutations for GitHub requests", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/settings/integrations",
      messages: [userMessage("Resync our GitHub repository.")],
    });

    expect(activeTools).toContain("resyncGitHubRepositoriesTool");
    expect(activeTools).not.toContain("bulkCreateStories");
    expect(activeTools.length).toBeLessThan(15);
  });

  it("falls back to a small read-oriented discovery set", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/inbox",
      messages: [userMessage("What needs my attention today?")],
    });

    expect(activeTools).toContain("focusBrief");
    expect(activeTools).not.toContain("deleteStory");
    expect(activeTools.length).toBeLessThan(10);
  });

  it.each([
    {
      excluded: "bulkDeleteStories",
      included: "teamPerformanceReportTool",
      path: "/acme/analytics",
      prompt: "Show team performance for Product.",
    },
    {
      excluded: "workspaceCommandCenterReportTool",
      included: "listCustomerFeedbackTool",
      path: "/acme/feedback",
      prompt: "Show customer feedback about onboarding.",
    },
    {
      excluded: "updateObjectiveTool",
      included: "listDocumentsTool",
      path: "/acme/documents",
      prompt: "Find the launch document.",
    },
  ])(
    "keeps $included requests compact and domain-specific",
    ({ excluded, included, path, prompt }) => {
      const activeTools = selectActiveTools({
        currentPath: path,
        messages: [userMessage(prompt)],
      });

      expect(activeTools).toContain(included);
      expect(activeTools).not.toContain(excluded);
      expect(activeTools.length).toBeLessThan(15);
    },
  );

  it("keeps only the exact pending mutation available for confirmation", () => {
    const pendingToolMessage = {
      id: "assistant-message",
      parts: [
        {
          input: { teamId: "team-1" },
          output: {
            message: "Please confirm the team update.",
            needsConfirmation: true,
            success: false,
          },
          state: "output-available",
          type: "tool-updateTeam",
        },
      ],
      role: "assistant",
    } as unknown as UIMessage;

    const activeTools = selectActiveTools({
      currentPath: "/acme/inbox",
      messages: [pendingToolMessage, userMessage("Approved.")],
    });

    expect(activeTools).toEqual(["suggestions", "updateTeam"]);
  });

  it("does not revive an older confirmation after the mutation completes", () => {
    const pendingToolMessage = {
      id: "assistant-pending",
      parts: [
        {
          input: { teamId: "team-1" },
          output: {
            message: "Please confirm the team update.",
            needsConfirmation: true,
            success: false,
          },
          state: "output-available",
          type: "tool-updateTeam",
        },
      ],
      role: "assistant",
    } as unknown as UIMessage;
    const completedToolMessage = {
      id: "assistant-completed",
      parts: [
        {
          input: { confirmed: true, teamId: "team-1" },
          output: { message: "Team updated.", success: true },
          state: "output-available",
          type: "tool-updateTeam",
        },
      ],
      role: "assistant",
    } as unknown as UIMessage;

    const activeTools = selectActiveTools({
      currentPath: "/acme/inbox",
      messages: [
        pendingToolMessage,
        userMessage("Approved.", "approval"),
        completedToolMessage,
        userMessage("What needs attention now?", "follow-up"),
      ],
    });

    expect(activeTools).toContain("focusBrief");
    expect(activeTools).not.toContain("updateTeam");
  });

  it("keeps a pending multi-action tool available for confirmation", () => {
    const pendingToolMessage = {
      id: "assistant-message",
      parts: [
        {
          input: { action: "create-label", name: "Customer" },
          output: {
            message: "Please confirm label creation.",
            needsConfirmation: true,
            success: false,
          },
          state: "output-available",
          type: "tool-labels",
        },
      ],
      role: "assistant",
    } as unknown as UIMessage;

    const activeTools = selectActiveTools({
      currentPath: "/acme/inbox",
      messages: [pendingToolMessage, userMessage("Approved.")],
    });

    expect(activeTools).toEqual(["suggestions", "labels"]);
  });

  it.each([
    ["Show recent workspace activity.", "activitySummaryTool"],
    ["Create a label for customer work.", "labels"],
    ["Delete this attachment.", "deleteAttachment"],
  ])("routes compact utility intent %s", (prompt, expectedTool) => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/inbox",
      messages: [userMessage(prompt)],
    });

    expect(activeTools).toContain(expectedTool);
    expect(activeTools.length).toBeLessThan(5);
  });

  it("uses the most specific current-page domain for short follow-ups", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/teams/team-1/objectives/objective-1",
      messages: [userMessage("What about this one?")],
    });

    expect(activeTools).toContain("getObjectiveDetailsTool");
    expect(activeTools).not.toContain("createTeamTool");
  });
});
