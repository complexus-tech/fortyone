/** @jest-environment node */ // eslint-disable-line tsdoc/syntax -- The AI SDK requires web streams.
/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { UIMessage } from "ai";
import { selectActiveTools } from "@/app/api/chat/active-tools";
import { MAYA_TOOL_ACTIONS } from "./tool-actions";
import { MAYA_TOOL_NAMES } from "./tool-names";
import {
  getMutationRoute,
  getMutationRoutesByDomainOperation,
  isMutationOperation,
  MUTATION_ROUTING_MANIFEST,
} from "./mutation-routing-manifest";
import {
  isMutationCapableToolName,
  MUTATING_ACTION_TOOL_NAMES,
  MUTATING_TOOL_ACTIONS,
  MUTATION_TOOL_NAMES,
  requiresMutationApproval,
} from "./tool-policy";

const NON_MATERIAL_TOOL_NAMES = [
  "navigation",
  "theme",
  "suggestions",
  "members",
  "resolveMember",
  "search",
  "workspacePerformanceReportTool",
  "workspaceCommandCenterReportTool",
  "pulseReportTool",
  "storyPerformanceReportTool",
  "objectiveProgressReportTool",
  "teamPerformanceReportTool",
  "sprintPerformanceReportTool",
  "timelineTrendsReportTool",
  "workloadPlanningTool",
  "focusBrief",
  "mayaWorkPlanTool",
  "activitySummaryTool",
  "getGitHubIntegrationTool",
  "getGitHubTeamSettingsTool",
  "getStoryGitHubLinksTool",
  "getStoryGitHubCommentsTool",
  "listIntegrationRequestsTool",
  "getIntegrationRequestTool",
  "getRequestGitHubCommentsTool",
  "listCustomerFeedbackTool",
  "getCustomerFeedbackTool",
  "listDocumentsTool",
  "getDocumentDetailsTool",
  "listLinkedGoogleFilesTool",
  "getLinkedGoogleFileContentTool",
  "listTeams",
  "listPublicTeams",
  "getTeamDetails",
  "listTeamMembers",
  "getTeamSettingsTool",
  "listTeamStories",
  "searchStories",
  "getStoryDetails",
  "listSprints",
  "listRunningSprints",
  "getSprintDetailsTool",
  "getSprintAnalyticsTool",
  "listKeyResultsTool",
  "getKeyResultActivitiesTool",
  "listAttachments",
  "listObjectivesTool",
  "listTeamObjectivesTool",
  "objectiveAnalyticsTool",
  "getObjectiveDetailsTool",
  "getObjectiveActivitiesTool",
  "storyActivities",
  "listMemories",
] as const;

const READ_ACTIONS_BY_TOOL = {
  comments: ["list-comments"],
  labels: ["list-labels"],
  links: ["list-links"],
  notifications: [
    "list-notifications",
    "get-unread-count",
    "filter-notifications",
  ],
  objectiveStatuses: [
    "list-objective-statuses",
    "get-objective-status-details",
  ],
  statuses: ["list-all-statuses", "list-team-statuses", "get-status-details"],
  storyLabels: ["get-story-labels"],
} as const satisfies Record<
  (typeof MUTATING_ACTION_TOOL_NAMES)[number],
  readonly string[]
>;

const REACHABILITY_CASES = [
  {
    intent: "What should I ask Maya next?",
    toolNames: ["suggestions"],
  },
  {
    intent:
      "Navigate to stories, switch to dark mode, and search for onboarding.",
    toolNames: ["navigation", "theme", "search"],
  },
  {
    intent: "Show every team, its members, details, and settings.",
    toolNames: [
      "members",
      "resolveMember",
      "listTeams",
      "listPublicTeams",
      "getTeamDetails",
      "listTeamMembers",
      "getTeamSettingsTool",
    ],
  },
  {
    intent: "Analyze overall workspace health and performance.",
    toolNames: [
      "workspacePerformanceReportTool",
      "workspaceCommandCenterReportTool",
    ],
  },
  {
    intent: "Show the pulse report.",
    toolNames: ["pulseReportTool"],
  },
  {
    intent: "Analyze story performance.",
    toolNames: ["storyPerformanceReportTool"],
  },
  {
    intent: "Analyze objective and key result progress.",
    toolNames: ["objectiveProgressReportTool"],
  },
  {
    intent: "Analyze team performance.",
    toolNames: ["teamPerformanceReportTool"],
  },
  {
    intent: "Analyze sprint performance.",
    toolNames: ["sprintPerformanceReportTool"],
  },
  {
    intent: "Analyze timeline trends.",
    toolNames: ["timelineTrendsReportTool"],
  },
  {
    intent: "Analyze workload capacity.",
    toolNames: ["workloadPlanningTool"],
  },
  {
    intent: "What should I focus on today?",
    toolNames: ["focusBrief"],
  },
  {
    intent: "Plan and schedule story PRO-142 on my calendar.",
    toolNames: ["mayaWorkPlanTool", "applyMayaWorkPlanTool"],
  },
  {
    intent: "Show recent workspace activity and what changed.",
    toolNames: ["activitySummaryTool"],
  },
  {
    intent: "Connect our GitHub repository integration.",
    toolNames: [
      "getGitHubIntegrationTool",
      "createGitHubInstallSessionTool",
      "getGitHubTeamSettingsTool",
      "getStoryGitHubLinksTool",
      "getStoryGitHubCommentsTool",
    ],
  },
  {
    intent: "Resync our GitHub repositories.",
    toolNames: ["resyncGitHubRepositoriesTool"],
  },
  {
    intent: "Create a GitHub issue sync link.",
    toolNames: ["createGitHubIssueSyncLinkTool"],
  },
  {
    intent: "Delete the GitHub issue sync link.",
    toolNames: ["deleteGitHubIssueSyncLinkTool"],
  },
  {
    intent: "Update our GitHub workspace settings.",
    toolNames: ["updateGitHubWorkspaceSettingsTool"],
  },
  {
    intent: "Update our GitHub team settings.",
    toolNames: ["updateGitHubTeamSettingsTool"],
  },
  {
    intent: "Post a GitHub comment on story PRO-142.",
    toolNames: ["postStoryGitHubCommentTool"],
  },
  {
    intent: "Remove the GitHub link from story PRO-142.",
    toolNames: ["deleteStoryGitHubLinkTool"],
  },
  {
    intent: "Update an integration request.",
    toolNames: [
      "listIntegrationRequestsTool",
      "getIntegrationRequestTool",
      "updateIntegrationRequestTool",
      "getRequestGitHubCommentsTool",
    ],
  },
  {
    intent: "Accept this integration request.",
    toolNames: ["acceptIntegrationRequestTool"],
  },
  {
    intent: "Decline this integration request.",
    toolNames: ["declineIntegrationRequestTool"],
  },
  {
    intent: "Accept all integration requests.",
    toolNames: ["acceptAllIntegrationRequestsTool"],
  },
  {
    intent: "Decline all integration requests.",
    toolNames: ["declineAllIntegrationRequestsTool"],
  },
  {
    intent: "Post a comment on this integration request.",
    toolNames: ["postRequestGitHubCommentTool"],
  },
  {
    intent: "Show customer feedback.",
    toolNames: ["listCustomerFeedbackTool", "getCustomerFeedbackTool"],
  },
  {
    intent: "Show our documents and document details.",
    toolNames: ["listDocumentsTool", "getDocumentDetailsTool"],
  },
  {
    intent: "Read the Google Drive files I explicitly selected for Maya.",
    toolNames: ["listLinkedGoogleFilesTool", "getLinkedGoogleFileContentTool"],
  },
  {
    intent: "Create or join a new Product team.",
    toolNames: ["createTeamTool", "joinTeam"],
  },
  { intent: "Update the Product team.", toolNames: ["updateTeam"] },
  {
    intent: "Delete or leave the Product team.",
    toolNames: ["deleteTeam", "leaveTeam"],
  },
  {
    intent: "Show stories in the Product backlog.",
    toolNames: [
      "listTeamStories",
      "searchStories",
      "getStoryDetails",
      "statuses",
    ],
  },
  {
    intent: "Create fifty launch stories.",
    toolNames: ["createStory", "bulkCreateStories"],
  },
  {
    intent: "Update and assign these stories to Joseph.",
    toolNames: ["updateStory", "bulkUpdateStories", "assignStoriesToUser"],
  },
  {
    intent: "Delete these stories.",
    toolNames: ["deleteStory", "bulkDeleteStories"],
  },
  { intent: "Duplicate story PRO-142.", toolNames: ["duplicateStory"] },
  { intent: "Restore deleted story PRO-142.", toolNames: ["restoreStory"] },
  {
    intent: "Remove the blocking association from story PRO-142.",
    toolNames: ["addStoryAssociation", "removeStoryAssociation"],
  },
  { intent: "Add a comment to PRO-142.", toolNames: ["comments"] },
  {
    intent: "Add a Customer label to PRO-142.",
    toolNames: ["labels", "storyLabels"],
  },
  { intent: "Show the activity for PRO-142.", toolNames: ["storyActivities"] },
  { intent: "Add a link to PRO-142.", toolNames: ["links"] },
  {
    intent: "Delete an attachment from PRO-142.",
    toolNames: ["listAttachments", "deleteAttachment"],
  },
  {
    intent: "Update sprint settings and show sprint details and analytics.",
    toolNames: [
      "listSprints",
      "listRunningSprints",
      "getSprintDetailsTool",
      "getSprintAnalyticsTool",
      "updateSprintSettings",
    ],
  },
  {
    intent:
      "Show objective details, analytics, activities, statuses, and key results.",
    toolNames: [
      "objectiveStatuses",
      "listKeyResultsTool",
      "getKeyResultActivitiesTool",
      "listObjectivesTool",
      "listTeamObjectivesTool",
      "objectiveAnalyticsTool",
      "getObjectiveDetailsTool",
      "getObjectiveActivitiesTool",
    ],
  },
  { intent: "Create an objective.", toolNames: ["createObjectiveTool"] },
  {
    intent: "Update the launch objective.",
    toolNames: ["updateObjectiveTool"],
  },
  {
    intent: "Delete the launch objective.",
    toolNames: ["deleteObjectiveTool"],
  },
  {
    intent: "Create a key result for the launch objective.",
    toolNames: ["createKeyResultTool"],
  },
  {
    intent: "Update the launch objective key result.",
    toolNames: ["updateKeyResultTool"],
  },
  {
    intent: "Delete the launch objective key result.",
    toolNames: ["deleteKeyResultTool"],
  },
  {
    intent: "Remember this launch rule.",
    toolNames: ["listMemories", "createMemory"],
  },
  { intent: "Update this memory.", toolNames: ["updateMemory"] },
  { intent: "Forget this memory.", toolNames: ["deleteMemory"] },
  { intent: "Show my notifications.", toolNames: ["notifications"] },
] as const;

const userMessage = (text: string): UIMessage => ({
  id: "user-message",
  parts: [{ text, type: "text" }],
  role: "user",
});

describe("Maya tool registry coverage", () => {
  it("routes every mutation-capable tool through one canonical manifest entry", () => {
    const mutationCapableToolNames = [
      ...MUTATION_TOOL_NAMES,
      ...MUTATING_ACTION_TOOL_NAMES,
    ];
    const manifestToolNames = MUTATION_ROUTING_MANIFEST.map(
      ({ toolName }) => toolName,
    );

    expect(new Set(manifestToolNames).size).toBe(manifestToolNames.length);
    expect([...manifestToolNames].sort()).toEqual(
      [...mutationCapableToolNames].sort(),
    );
    mutationCapableToolNames.forEach((toolName) => {
      expect(
        manifestToolNames.filter((name) => name === toolName),
      ).toHaveLength(1);
    });

    for (const entry of MUTATION_ROUTING_MANIFEST) {
      expect(getMutationRoute(entry.toolName)).toBe(entry);
      expect(entry.operations.length).toBeGreaterThan(0);
      expect(new Set(entry.operations).size).toBe(entry.operations.length);
    }

    expect(
      getMutationRoutesByDomainOperation("story", "create").map(
        ({ toolName }) => toolName,
      ),
    ).toEqual(["createStory", "bulkCreateStories"]);
    expect(
      getMutationRoutesByDomainOperation("comment", "add-comment").map(
        ({ toolName }) => toolName,
      ),
    ).toEqual(["comments"]);
    expect(isMutationOperation("add-comment")).toBe(true);
    expect(isMutationOperation("list-comments")).toBe(false);
  });

  it("routes every registered tool from at least one clear natural-language intent", () => {
    const coveredToolNames = new Set(
      REACHABILITY_CASES.flatMap(({ toolNames }) => toolNames),
    );

    expect(Array.from(coveredToolNames).sort()).toEqual(
      [...MAYA_TOOL_NAMES].sort(),
    );

    for (const { intent, toolNames } of REACHABILITY_CASES) {
      const activeTools = selectActiveTools({
        currentPath: "/acme/maya",
        messages: [userMessage(intent)],
      });
      expect(activeTools).toEqual(expect.arrayContaining([...toolNames]));
    }
  });

  it("requires every registered tool to have an explicit mutation classification", () => {
    const classifiedToolNames = new Set([
      ...NON_MATERIAL_TOOL_NAMES,
      ...MUTATION_TOOL_NAMES,
      ...MUTATING_ACTION_TOOL_NAMES,
    ]);

    expect(Array.from(classifiedToolNames).sort()).toEqual(
      [...MAYA_TOOL_NAMES].sort(),
    );
    NON_MATERIAL_TOOL_NAMES.forEach((toolName) => {
      expect(isMutationCapableToolName(toolName)).toBe(false);
    });
    MUTATION_TOOL_NAMES.forEach((toolName) => {
      expect(isMutationCapableToolName(toolName)).toBe(true);
    });
    MUTATING_ACTION_TOOL_NAMES.forEach((toolName) => {
      expect(isMutationCapableToolName(toolName)).toBe(true);
    });
  });

  it("keeps every action-scoped operation in an explicit approval policy", () => {
    expect(Object.keys(MAYA_TOOL_ACTIONS).sort()).toEqual(
      [...MUTATING_ACTION_TOOL_NAMES].sort(),
    );

    for (const toolName of MUTATING_ACTION_TOOL_NAMES) {
      const mutatingActions = Array.from(MUTATING_TOOL_ACTIONS[toolName]);
      const readActions = READ_ACTIONS_BY_TOOL[toolName];

      expect([...MAYA_TOOL_ACTIONS[toolName]].sort()).toEqual(
        [...mutatingActions, ...readActions].sort(),
      );
      mutatingActions.forEach((action) => {
        expect(requiresMutationApproval(toolName, { action })).toBe(true);
      });
      readActions.forEach((action) => {
        expect(requiresMutationApproval(toolName, { action })).toBe(false);
      });
    }
  });

  it("requires native approval for every standalone material mutation", () => {
    MUTATION_TOOL_NAMES.forEach((toolName) => {
      expect(requiresMutationApproval(toolName, {})).toBe(
        toolName !== "createGitHubInstallSessionTool",
      );
    });
  });
});
