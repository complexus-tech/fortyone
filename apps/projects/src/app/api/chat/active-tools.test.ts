/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { UIMessage } from "ai";
import { selectActiveTools } from "./active-tools";

const userMessage = (text: string, id = "user-message") =>
  ({
    id,
    parts: [{ text, type: "text" }],
    role: "user",
  }) as UIMessage;

const assistantToolMessage = ({
  id = "assistant-message",
  output,
  toolName,
}: {
  id?: string;
  output: unknown;
  toolName: string;
}) =>
  ({
    id,
    parts: [
      {
        input: {},
        output,
        state: "output-available",
        type: `tool-${toolName}`,
      },
    ],
    role: "assistant",
  }) as unknown as UIMessage;

const assistantTextMessage = (text: string, id = "assistant-message") =>
  ({
    id,
    parts: [{ text, type: "text" }],
    role: "assistant",
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

  it.each(["ticket", "tickets", "issue", "issues"])(
    "treats %s as story terminology",
    (storyTerm) => {
      const activeTools = selectActiveTools({
        currentPath: "/acme/inbox",
        messages: [userMessage(`Create two launch ${storyTerm}.`)],
      });

      expect(activeTools).toEqual(
        expect.arrayContaining(["createStory", "bulkCreateStories"]),
      );
      expect(activeTools).not.toContain("createTeamTool");
    },
  );

  it("recognizes approval-recovery wording as story creation", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/inbox",
      messages: [userMessage("Prepare the same 50 tickets again now.")],
    });

    expect(activeTools).toEqual(
      expect.arrayContaining(["createStory", "bulkCreateStories"]),
    );
  });

  it("uses safely normalized workspace story terminology", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/inbox",
      messages: [userMessage("Create two launch work items.")],
      storyTerminology: "Work Items",
    });

    expect(activeTools).toEqual(
      expect.arrayContaining(["createStory", "bulkCreateStories"]),
    );
  });

  it("carries workspace story terminology through a confirmation follow-up", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/inbox",
      messages: [
        assistantTextMessage(
          "I can create two launch work items. Please confirm.",
        ),
        userMessage("Yes, do it."),
      ],
      storyTerminology: "work items",
    });

    expect(activeTools).toEqual(
      expect.arrayContaining(["createStory", "bulkCreateStories"]),
    );
  });

  it("keeps story creation tools available for a duration-only planning reply", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/maya",
      messages: [
        userMessage("Create a launch story for the Product team.", "request"),
        assistantTextMessage(
          "How much time is needed for this story?",
          "clarification",
        ),
        userMessage("60 minutes", "planning-reply"),
      ],
    });

    expect(activeTools).toEqual(
      expect.arrayContaining(["createStory", "bulkCreateStories"]),
    );
    expect(activeTools).not.toContain("mayaWorkPlanTool");
  });

  it("routes a complete terse planning reply back to story creation", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/maya",
      messages: [
        userMessage("Create a launch task for the Product team.", "request"),
        assistantTextMessage(
          "When should the task be delivered, how much time is needed, and should calendar scheduling be on?",
          "clarification",
        ),
        userMessage("Friday, 60 minutes, scheduling off", "planning-reply"),
      ],
    });

    expect(activeTools).toEqual(
      expect.arrayContaining(["createStory", "bulkCreateStories"]),
    );
    expect(activeTools).not.toContain("mayaWorkPlanTool");
    expect(activeTools).not.toContain("applyMayaWorkPlanTool");
  });

  it("keeps bulk creation available for per-story planning details", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/maya",
      messages: [
        userMessage(
          "Create three launch tickets: API, web, and docs.",
          "request",
        ),
        assistantTextMessage(
          "These tickets will stay unscheduled by default. Please provide any per-ticket delivery dates and time needed that you want me to use.",
          "clarification",
        ),
        userMessage(
          "API: Friday, 45 minutes. Web: Monday, 2 hours. Leave docs unscheduled.",
          "planning-reply",
        ),
      ],
    });

    expect(activeTools).toContain("bulkCreateStories");
    expect(activeTools).not.toContain("bulkDeleteStories");
    expect(activeTools).not.toContain("mayaWorkPlanTool");
  });

  it("does not treat an unrelated duration answer as story creation", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/maya",
      messages: [
        userMessage("Help me plan a meeting.", "request"),
        assistantTextMessage(
          "How much time do you want for the meeting?",
          "clarification",
        ),
        userMessage("60 minutes", "planning-reply"),
      ],
    });

    expect(activeTools).not.toContain("createStory");
    expect(activeTools).not.toContain("bulkCreateStories");
  });

  it("ignores invalid custom terminology instead of treating it as a pattern", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/inbox",
      messages: [userMessage("Create a launch plan.")],
      storyTerminology: ".*",
    });

    expect(activeTools).not.toContain("createStory");
    expect(activeTools).not.toContain("bulkCreateStories");
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

  it("uses story tool provenance for a pronoun-based delete follow-up", () => {
    const completedBulkCreate = assistantToolMessage({
      output: {
        createdStories: [{ id: "story-1" }, { id: "story-2" }],
        success: true,
      },
      toolName: "bulkCreateStories",
    });

    const activeTools = selectActiveTools({
      currentPath: "/acme/inbox",
      messages: [
        userMessage("Create two launch tasks.", "create-request"),
        completedBulkCreate,
        userMessage("Let's delete them.", "delete-follow-up"),
      ],
    });

    expect(activeTools).toEqual(
      expect.arrayContaining(["deleteStory", "bulkDeleteStories"]),
    );
    expect(activeTools).not.toContain("createStory");
    expect(activeTools).not.toContain("bulkCreateStories");
  });

  it("routes noun-form deletion requests to the story deletion tools", () => {
    const activeTools = selectActiveTools({
      currentPath: "/first/maya",
      messages: [
        userMessage(
          "Prepare both deletion actions using the actual deletion tools for the 50 QA tickets. Do not execute either action.",
        ),
      ],
    });

    expect(activeTools).toEqual(
      expect.arrayContaining(["deleteStory", "bulkDeleteStories"]),
    );
    expect(activeTools).not.toContain("createStory");
    expect(activeTools).not.toContain("bulkCreateStories");
  });

  it.each([
    "Mark PRO-142 done.",
    "Complete these tickets.",
    "Close the launch story.",
    "Finish task PRO-142.",
    "Reopen PRO-142.",
  ])("routes common story status language to update tools: %s", (prompt) => {
    const activeTools = selectActiveTools({
      currentPath: "/first/maya",
      messages: [userMessage(prompt)],
    });

    expect(activeTools).toEqual(
      expect.arrayContaining(["updateStory", "bulkUpdateStories"]),
    );
  });

  it("does not carry the previous story action into a neutral follow-up", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/inbox",
      messages: [
        assistantToolMessage({
          output: { story: { id: "story-1" }, success: true },
          toolName: "createStory",
        }),
        userMessage("What about that one?"),
      ],
    });

    expect(activeTools).toContain("getStoryDetails");
    expect(activeTools).not.toContain("createStory");
    expect(activeTools).not.toContain("updateStory");
    expect(activeTools).not.toContain("deleteStory");
  });

  it.each(["Yes, do it.", "Do everything except changing the sprint."])(
    "uses a text-only story proposal as domain context for: %s",
    (followUp) => {
      const activeTools = selectActiveTools({
        currentPath: "/acme/inbox",
        messages: [
          userMessage("Help me plan the launch.", "planning-request"),
          assistantTextMessage(
            "I can create eight launch stories and update the existing report story. Please confirm.",
          ),
          userMessage(followUp, "confirmation"),
        ],
      });

      expect(activeTools).toEqual(
        expect.arrayContaining(["bulkCreateStories", "bulkUpdateStories"]),
      );
      expect(activeTools).not.toContain("bulkDeleteStories");
    },
  );

  it.each([
    {
      expected: "deleteObjectiveTool",
      excluded: "deleteStory",
      followUp: "Delete it.",
      toolName: "createObjectiveTool",
    },
    {
      expected: "deleteTeam",
      excluded: "deleteObjectiveTool",
      followUp: "Delete that.",
      toolName: "createTeamTool",
    },
    {
      expected: "updateMemory",
      excluded: "updateStory",
      followUp: "Update that.",
      toolName: "createMemory",
    },
    {
      expected: "acceptIntegrationRequestTool",
      excluded: "bulkUpdateStories",
      followUp: "Accept them.",
      toolName: "listIntegrationRequestsTool",
    },
    {
      expected: "notifications",
      excluded: "bulkUpdateStories",
      followUp: "Mark those as read.",
      toolName: "notifications",
    },
    {
      expected: "resyncGitHubRepositoriesTool",
      excluded: "bulkUpdateStories",
      followUp: "Resync it.",
      toolName: "getGitHubIntegrationTool",
    },
  ])(
    "carries $toolName domain provenance into a conversational follow-up",
    ({ excluded, expected, followUp, toolName }) => {
      const activeTools = selectActiveTools({
        currentPath: "/acme/inbox",
        messages: [
          assistantToolMessage({
            output: { success: true },
            toolName,
          }),
          userMessage(followUp),
        ],
      });

      expect(activeTools).toContain(expected);
      expect(activeTools).not.toContain(excluded);
    },
  );

  it("does not reuse a mutation verb from assistant text for a neutral follow-up", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/inbox",
      messages: [
        assistantTextMessage(
          "I can delete those launch stories after you confirm.",
        ),
        userMessage("What about them?"),
      ],
    });

    expect(activeTools).toContain("getStoryDetails");
    expect(activeTools).not.toContain("createStory");
    expect(activeTools).not.toContain("updateStory");
    expect(activeTools).not.toContain("deleteStory");
  });

  it("keeps work-plan tools available when a user accepts a contextual nudge", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/inbox",
      messages: [
        assistantTextMessage(
          "This work item has no owner and a tight delivery window. I can create a Maya work plan to assign it and protect calendar time.",
        ),
        userMessage("Yes, do it."),
      ],
      storyTerminology: "work item",
    });

    expect(activeTools).toEqual(
      expect.arrayContaining(["mayaWorkPlanTool", "applyMayaWorkPlanTool"]),
    );
  });

  it.each([
    "Analyze my workspace.",
    "Give me a workspace analysis.",
    "Show workspace insights and health.",
    "Give me a workspace overview.",
    "Analyze workspace health across all teams.",
  ])(
    "routes workspace analysis language through analytics tools: %s",
    (prompt) => {
      const activeTools = selectActiveTools({
        currentPath: "/acme/inbox",
        messages: [userMessage(prompt)],
      });

      expect(activeTools).toContain("workspacePerformanceReportTool");
      expect(activeTools).toContain("workspaceCommandCenterReportTool");
      expect(activeTools).toContain("listTeams");
      expect(activeTools).not.toContain("bulkDeleteStories");
    },
  );

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

  it("routes a story-reference activity request to the detailed timeline", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/inbox",
      messages: [userMessage("Show me the activity for PRO-142.")],
    });

    expect(activeTools).toEqual(
      expect.arrayContaining([
        "searchStories",
        "getStoryDetails",
        "storyActivities",
      ]),
    );
    expect(activeTools).not.toContain("activitySummaryTool");
  });

  it("uses the current story page for a detail activity request", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/story/story-uuid/launch-checklist",
      messages: [userMessage("Show its recent activity.")],
    });

    expect(activeTools).toContain("storyActivities");
    expect(activeTools).not.toContain("activitySummaryTool");
  });

  it("routes team status creation without exposing team creation", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/inbox",
      messages: [
        userMessage("Create a Ready for QA status for the Product team."),
      ],
    });

    expect(activeTools).toEqual(
      expect.arrayContaining(["statuses", "listTeams"]),
    );
    expect(activeTools).not.toContain("createTeamTool");
    expect(activeTools).not.toContain("joinTeam");
  });

  it.each(["Add the Customer label to PRO-142.", "Tag PRO-142 as Customer."])(
    "routes a story label request by reference: %s",
    (prompt) => {
      const activeTools = selectActiveTools({
        currentPath: "/acme/inbox",
        messages: [userMessage(prompt)],
      });

      expect(activeTools).toEqual(
        expect.arrayContaining(["searchStories", "labels", "storyLabels"]),
      );
    },
  );

  it("uses the most specific current-page domain for short follow-ups", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/teams/team-1/objectives/objective-1",
      messages: [userMessage("What about this one?")],
    });

    expect(activeTools).toContain("getObjectiveDetailsTool");
    expect(activeTools).not.toContain("createTeamTool");
  });
});
