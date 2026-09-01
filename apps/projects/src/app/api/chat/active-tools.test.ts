/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { UIMessage } from "ai";
import { selectActiveTools } from "./active-tools";

jest.mock("server-only", () => ({}));

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

  it("keeps story creation available throughout a multi-turn intake", () => {
    const messages = [
      userMessage("create a story", "create-request"),
      assistantTextMessage(
        "What should the story be called? I’ll create it in the Product team. Also provide a due/work date and time needed, or say skip planning; no calendar time will be reserved unless you explicitly request it.",
        "title-clarification",
      ),
      userMessage("test ai", "title-reply"),
      assistantTextMessage(
        "I’ll create “Test AI” in the Product team. What due/work date and time needed should I use, or should I skip planning?",
        "planning-clarification",
      ),
      userMessage("skip planning assign to me", "planning-reply"),
    ];

    const afterTitleReply = selectActiveTools({
      currentPath: "/acme/maya",
      messages: messages.slice(0, 3),
    });
    const afterPlanningReply = selectActiveTools({
      currentPath: "/acme/maya",
      messages,
    });

    for (const activeTools of [afterTitleReply, afterPlanningReply]) {
      expect(activeTools).toEqual(
        expect.arrayContaining(["createStory", "bulkCreateStories"]),
      );
      expect(activeTools).not.toContain("bulkDeleteStories");
      expect(activeTools).not.toContain("mayaWorkPlanTool");
    }
  });

  it("carries story creation through consecutive team and title clarifications", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/maya",
      messages: [
        userMessage("Create a task.", "create-request"),
        assistantTextMessage(
          "Which team should I create the task in?",
          "team-clarification",
        ),
        userMessage("Product", "team-reply"),
        assistantTextMessage(
          "What should the task be called?",
          "title-clarification",
        ),
        userMessage("Retire the legacy API", "title-reply"),
      ],
    });

    expect(activeTools).toEqual(
      expect.arrayContaining(["createStory", "bulkCreateStories"]),
    );
    expect(activeTools).not.toContain("bulkDeleteStories");
  });

  it.each([
    "I’ll create “Create a team dashboard” in the Product team. What due/work date and time needed should I use?",
    "I’ll create 'Create a team dashboard' in the Product team. What due/work date and time needed should I use?",
    "I’ll create **Create a team dashboard** in the Product team. What due/work date and time needed should I use?",
  ])(
    "treats a command-shaped title as an intake value when planning continues: %s",
    (planningClarification) => {
      const activeTools = selectActiveTools({
        currentPath: "/acme/maya",
        messages: [
          userMessage("Create a story.", "create-request"),
          assistantTextMessage(
            "What should the story be called?",
            "title-clarification",
          ),
          userMessage("Create a team dashboard", "title-reply"),
          assistantTextMessage(planningClarification, "planning-clarification"),
          userMessage("skip planning", "planning-reply"),
        ],
      });

      expect(activeTools).toEqual(
        expect.arrayContaining(["createStory", "bulkCreateStories"]),
      );
      expect(activeTools).not.toContain("bulkDeleteStories");
      expect(activeTools).not.toContain("mayaWorkPlanTool");
    },
  );

  it("carries story creation through consecutive generic slot prompts", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/maya",
      messages: [
        userMessage("Create a story.", "create-request"),
        assistantTextMessage(
          "What should it be called?",
          "title-clarification",
        ),
        userMessage("Launch dashboard", "title-reply"),
        assistantTextMessage("Which team?", "team-clarification"),
        userMessage("Product", "team-reply"),
        assistantTextMessage("Which status?", "status-clarification"),
        userMessage("Backlog", "status-reply"),
      ],
    });

    expect(activeTools).toEqual(
      expect.arrayContaining(["createStory", "bulkCreateStories"]),
    );
    expect(activeTools).not.toContain("createTeamTool");
  });

  it("keeps story creation available through a long intake chain", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/maya",
      messages: [
        userMessage("Create a story.", "create-request"),
        assistantTextMessage(
          "Which team should I create the story in?",
          "team-clarification",
        ),
        userMessage("Product", "team-reply"),
        assistantTextMessage(
          "What should the story be called?",
          "title-clarification",
        ),
        userMessage("Launch dashboard", "title-reply"),
        assistantTextMessage(
          "Which status should the story use?",
          "status-clarification",
        ),
        userMessage("Backlog", "status-reply"),
        assistantTextMessage(
          "Who should be assigned to the story?",
          "assignee-clarification",
        ),
        userMessage("Assign it to me", "assignee-reply"),
        assistantTextMessage(
          "What priority should the story use?",
          "priority-clarification",
        ),
        userMessage("High priority", "priority-reply"),
        assistantTextMessage(
          "What due/work date and time needed should I use for the story?",
          "planning-clarification",
        ),
        userMessage("Skip planning", "planning-reply"),
      ],
    });

    expect(activeTools).toEqual(
      expect.arrayContaining(["createStory", "bulkCreateStories"]),
    );
    expect(activeTools).not.toContain("bulkUpdateStories");
    expect(activeTools).not.toContain("mayaWorkPlanTool");
  });

  it("does not treat a generic naming clarification as story creation", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/maya",
      messages: [
        userMessage("Help me configure the integration.", "request"),
        assistantTextMessage(
          "What should the connection be called?",
          "clarification",
        ),
        userMessage("Test AI", "reply"),
      ],
    });

    expect(activeTools).not.toContain("createStory");
    expect(activeTools).not.toContain("bulkCreateStories");
  });

  it("does not carry an older story intake across a generic domain flow", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/maya",
      messages: [
        userMessage("Create a story.", "story-request"),
        assistantTextMessage(
          "What should the story be called?",
          "story-clarification",
        ),
        userMessage(
          "Help me configure the integration.",
          "integration-request",
        ),
        assistantTextMessage(
          "What should it be called?",
          "integration-clarification",
        ),
        userMessage("Linear", "integration-name"),
      ],
    });

    expect(activeTools).not.toContain("createStory");
    expect(activeTools).not.toContain("bulkCreateStories");
  });

  it("does not revive an earlier story intake after a creation redirection", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/maya",
      messages: [
        userMessage("Create a story.", "story-request"),
        assistantTextMessage(
          "What should the story be called?",
          "story-clarification",
        ),
        userMessage("Actually, create a team instead.", "creation-redirection"),
        assistantTextMessage(
          "I’ll create the team. What should it be called?",
          "team-clarification",
        ),
        userMessage("Platform", "team-name"),
      ],
    });

    expect(activeTools).not.toContain("createStory");
    expect(activeTools).not.toContain("bulkCreateStories");
  });

  it("honors a direct creation redirection from a story clarification", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/maya",
      messages: [
        userMessage("Create a story.", "story-request"),
        assistantTextMessage(
          "What should the story be called?",
          "story-clarification",
        ),
        userMessage("Actually, create a team instead.", "creation-redirection"),
      ],
    });

    expect(activeTools).toContain("createTeamTool");
    expect(activeTools).not.toContain("createStory");
    expect(activeTools).not.toContain("bulkCreateStories");
  });

  it.each([
    ["Actually add a comment instead", ["comments"]],
    ["Actually add a link instead", ["links"]],
    [
      "Actually make a work plan instead",
      ["mayaWorkPlanTool", "applyMayaWorkPlanTool"],
    ],
  ] as const)(
    "routes an explicit non-story redirection: %s",
    (redirection, expectedTools) => {
      const activeTools = selectActiveTools({
        currentPath: "/acme/maya",
        messages: [
          userMessage("Create a story.", "story-request"),
          assistantTextMessage(
            "What should the story be called?",
            "story-clarification",
          ),
          userMessage(redirection, "domain-redirection"),
        ],
      });

      expect(activeTools).toEqual(expect.arrayContaining(expectedTools));
      expect(activeTools).not.toContain("createStory");
      expect(activeTools).not.toContain("bulkCreateStories");
    },
  );

  it.each([
    ["Create a team", ["createTeamTool"]],
    ["Help me configure the integration", []],
    ["Add a comment", ["comments"]],
    ["Add a link", ["links"]],
    ["Make a work plan", ["mayaWorkPlanTool", "applyMayaWorkPlanTool"]],
  ] as const)(
    "honors a decisive non-story request without a redirection cue: %s",
    (request, expectedTools) => {
      const activeTools = selectActiveTools({
        currentPath: "/acme/maya",
        messages: [
          userMessage("Create a story.", "story-request"),
          assistantTextMessage(
            "What should the story be called?",
            "story-clarification",
          ),
          userMessage(request, "domain-request"),
        ],
      });

      expect(activeTools).toEqual(expect.arrayContaining(expectedTools));
      expect(activeTools).not.toContain("createStory");
      expect(activeTools).not.toContain("bulkCreateStories");
    },
  );

  it.each([
    ["Who should be assigned?", "Actually assign it to me instead"],
    ["Which status?", "Actually set the status to Done instead"],
    ["Which team?", "Actually change the team to Product instead"],
    ["When should the story be scheduled?", "Schedule it"],
    ["Which objective should it link to?", "Link it to Launch objective"],
  ])(
    "keeps metadata corrections inside story creation: %s → %s",
    (clarification, correction) => {
      const activeTools = selectActiveTools({
        currentPath: "/acme/maya",
        messages: [
          userMessage("Create a story.", "story-request"),
          assistantTextMessage(clarification, "metadata-clarification"),
          userMessage(correction, "metadata-correction"),
        ],
      });

      expect(activeTools).toEqual(
        expect.arrayContaining(["createStory", "bulkCreateStories"]),
      );
      expect(activeTools).not.toContain("updateStory");
      expect(activeTools).not.toContain("bulkUpdateStories");
    },
  );

  it.each([
    [
      "What due/work date and time needed should I use for the story?",
      "Never mind the planning; just create it and assign it to me",
    ],
    [
      "What should the story title be?",
      "Forget that title; call it Launch v2 instead",
    ],
  ])(
    "treats cancellation wording with a replacement as an intake correction: %s → %s",
    (clarification, correction) => {
      const activeTools = selectActiveTools({
        currentPath: "/acme/maya",
        messages: [
          userMessage("Create a story.", "story-request"),
          assistantTextMessage(clarification, "intake-clarification"),
          userMessage(correction, "intake-correction"),
        ],
      });

      expect(activeTools).toEqual(
        expect.arrayContaining(["createStory", "bulkCreateStories"]),
      );
      expect(activeTools).not.toContain("updateStory");
      expect(activeTools).not.toContain("bulkUpdateStories");
    },
  );

  it("ends story intake for a bare cancellation", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/maya",
      messages: [
        userMessage("Create a story.", "story-request"),
        assistantTextMessage(
          "What due/work date and time needed should I use for the story?",
          "planning-clarification",
        ),
        userMessage("Never mind", "cancellation"),
      ],
    });

    expect(activeTools).not.toContain("createStory");
    expect(activeTools).not.toContain("bulkCreateStories");
  });

  it("stops at a newer non-story creation request without explicit redirection", () => {
    const activeTools = selectActiveTools({
      currentPath: "/acme/maya",
      messages: [
        userMessage("Create a story.", "story-request"),
        assistantTextMessage(
          "What should the story be called?",
          "story-clarification",
        ),
        userMessage("Create a team.", "team-request"),
        assistantTextMessage("What should it be called?", "team-clarification"),
        userMessage("Platform", "team-name"),
      ],
    });

    expect(activeTools).not.toContain("createStory");
    expect(activeTools).not.toContain("bulkCreateStories");
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
