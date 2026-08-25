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
  });

  it("falls back to a small read-oriented discovery set", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/inbox",
      messages: [userMessage("What needs my attention today?")],
    });

    expect(activeTools).toContain("focusBrief");
    expect(activeTools).not.toContain("deleteStory");
    expect(activeTools.length).toBeLessThan(25);
  });
});
