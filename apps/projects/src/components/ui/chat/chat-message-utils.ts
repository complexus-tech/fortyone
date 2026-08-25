import type { MayaUIMessage } from "@/lib/ai/tools/types";
import type { ToolMessagePart } from "./tool-output-policy";
import {
  isAnalyticsReportOutput,
  isMemberResolverToolPart,
  isMutationToolPart,
  isRenderableToolPart,
  isSupportingToolType,
  isToolMessagePart,
} from "./tool-output-policy";

const TEAM_SCOPED_RESULT_TYPES = new Set([
  "tool-listTeamStories",
  "tool-listSprints",
  "tool-listRunningSprints",
  "tool-listTeamObjectivesTool",
  "tool-listTeamMembers",
  "tool-listCustomerFeedbackTool",
  "tool-listIntegrationRequestsTool",
]);

const TEAM_RESOLVER_TYPES = new Set(["tool-listTeams", "tool-listPublicTeams"]);

const MEMBER_RESULT_TYPES = new Set(["tool-members", "tool-listTeamMembers"]);

const MEMBER_SCOPED_RESULT_TYPES = new Set([
  "tool-listTeamStories",
  "tool-searchStories",
  "tool-focusBrief",
  "tool-workspacePerformanceReportTool",
  "tool-workspaceCommandCenterReportTool",
  "tool-pulseReportTool",
  "tool-storyPerformanceReportTool",
  "tool-objectiveProgressReportTool",
  "tool-teamPerformanceReportTool",
  "tool-sprintPerformanceReportTool",
  "tool-timelineTrendsReportTool",
  "tool-workloadPlanningTool",
  "tool-mayaWorkPlanTool",
  "tool-activitySummaryTool",
]);

const PROMPT_URL_PATTERN = /\b(?:https?:\/\/|www\.)[^\s<>"']+/gi;
const TRAILING_PROMPT_URL_PUNCTUATION = /[.,!?;:]+$/;

export type PromptTextSegment =
  | { type: "link"; href: string; start: number; value: string }
  | { type: "text"; start: number; value: string };

export const getPromptTextSegments = (text: string): PromptTextSegment[] => {
  const segments: PromptTextSegment[] = [];
  const urlPattern = new RegExp(PROMPT_URL_PATTERN.source, "gi");
  let previousIndex = 0;

  let match = urlPattern.exec(text);

  while (match) {
    const matchIndex = match.index;
    const rawValue = match[0];
    const trailingPunctuation =
      TRAILING_PROMPT_URL_PUNCTUATION.exec(rawValue)?.[0] ?? "";
    const value = rawValue.slice(
      0,
      trailingPunctuation ? -trailingPunctuation.length : undefined,
    );

    if (matchIndex > previousIndex) {
      segments.push({
        start: previousIndex,
        type: "text",
        value: text.slice(previousIndex, matchIndex),
      });
    }

    segments.push({
      type: "link",
      href: value.startsWith("www.") ? `https://${value}` : value,
      start: matchIndex,
      value,
    });

    if (trailingPunctuation) {
      segments.push({
        start: matchIndex + value.length,
        type: "text",
        value: trailingPunctuation,
      });
    }

    previousIndex = matchIndex + rawValue.length;
    match = urlPattern.exec(text);
  }

  if (previousIndex < text.length) {
    segments.push({
      start: previousIndex,
      type: "text",
      value: text.slice(previousIndex),
    });
  }

  return segments.length > 0
    ? segments
    : [{ start: 0, type: "text", value: text }];
};

export const getVisibleToolPartIndexes = (message: MayaUIMessage) => {
  const invokedToolParts = message.parts.flatMap((part, index) =>
    isToolMessagePart(part) ? [{ index, part }] : [],
  );
  const mutationPartIndexes = message.parts.flatMap((part, index) =>
    isToolMessagePart(part) &&
    part.state === "output-available" &&
    isMutationToolPart(part)
      ? [index]
      : [],
  );
  const toolParts = message.parts.flatMap((part, index) =>
    isToolMessagePart(part) && isRenderableToolPart(part)
      ? [{ index, part }]
      : [],
  );

  return new Set(
    toolParts.flatMap(({ index, part }, toolIndex) => {
      const laterToolParts = toolParts.slice(toolIndex + 1);
      const laterInvokedToolParts = invokedToolParts.filter(
        ({ index: laterIndex }) => laterIndex > index,
      );
      const feedsLaterReport =
        !isAnalyticsReportOutput(part.output) &&
        laterToolParts.some(({ part: laterPart }) =>
          isAnalyticsReportOutput(laterPart.output),
        );
      const feedsLaterTeamResult =
        TEAM_RESOLVER_TYPES.has(part.type) &&
        laterToolParts.some(({ part: laterPart }) =>
          TEAM_SCOPED_RESULT_TYPES.has(laterPart.type),
        );
      const feedsLaterMemberResult =
        (isMemberResolverToolPart(part) ||
          (MEMBER_RESULT_TYPES.has(part.type) && !("input" in part))) &&
        laterInvokedToolParts.some(
          ({ part: laterPart }) =>
            MEMBER_SCOPED_RESULT_TYPES.has(laterPart.type) ||
            isMutationToolPart(laterPart),
        );
      const hasLaterResultOfSameType = laterToolParts.some(
        ({ part: laterPart }) => laterPart.type === part.type,
      );
      const feedsLaterMutation =
        !isMutationToolPart(part) &&
        mutationPartIndexes.some(
          (mutationPartIndex) => mutationPartIndex > index,
        );

      return feedsLaterReport ||
        feedsLaterTeamResult ||
        feedsLaterMemberResult ||
        hasLaterResultOfSameType ||
        feedsLaterMutation
        ? []
        : [index];
    }),
  );
};

/** Maps tool part types to the single progress label shown below the chat. */
const TOOL_THINKING_LABELS: Record<string, string> = {
  // Stories
  "tool-listTeamStories": "Fetching stories",
  "tool-searchStories": "Searching stories",
  "tool-getStoryDetails": "Getting story details",
  "tool-createStory": "Creating story",
  "tool-updateStory": "Updating story",
  "tool-deleteStory": "Deleting story",
  "tool-bulkCreateStories": "Creating stories",
  "tool-bulkUpdateStories": "Updating stories",
  "tool-bulkDeleteStories": "Deleting stories",
  "tool-assignStoriesToUser": "Assigning stories",
  "tool-duplicateStory": "Duplicating story",
  "tool-restoreStory": "Restoring story",
  "tool-addStoryAssociation": "Linking stories",
  "tool-removeStoryAssociation": "Unlinking stories",
  // Sprints
  "tool-listSprints": "Loading sprints",
  "tool-listRunningSprints": "Getting active sprints",
  "tool-getSprintDetailsTool": "Getting sprint details",
  "tool-getSprintAnalyticsTool": "Analyzing sprint data",
  "tool-updateSprintSettings": "Updating sprint settings",
  // Teams
  "tool-listTeams": "Loading teams",
  "tool-listPublicTeams": "Loading public teams",
  "tool-getTeamDetails": "Getting team details",
  "tool-listTeamMembers": "Loading team members",
  "tool-createTeamTool": "Creating team",
  "tool-updateTeam": "Updating team",
  "tool-joinTeam": "Joining team",
  "tool-leaveTeam": "Leaving team",
  "tool-deleteTeam": "Deleting team",
  "tool-getTeamSettingsTool": "Loading team settings",
  // Objectives & Key Results
  "tool-listObjectivesTool": "Loading objectives",
  "tool-listTeamObjectivesTool": "Loading team objectives",
  "tool-createObjectiveTool": "Creating objective",
  "tool-updateObjectiveTool": "Updating objective",
  "tool-deleteObjectiveTool": "Deleting objective",
  "tool-getObjectiveDetailsTool": "Getting objective details",
  "tool-objectiveAnalyticsTool": "Analyzing objective data",
  "tool-getObjectiveActivitiesTool": "Loading objective activity",
  "tool-listKeyResultsTool": "Loading key results",
  "tool-createKeyResultTool": "Creating key result",
  "tool-updateKeyResultTool": "Updating key result",
  "tool-deleteKeyResultTool": "Deleting key result",
  "tool-getKeyResultActivitiesTool": "Loading key result activity",
  // Other
  "tool-navigation": "Navigating",
  "tool-search": "Searching",
  "tool-members": "Loading members",
  "tool-resolveMember": "Resolving member",
  "tool-comments": "Loading comments",
  "tool-notifications": "Checking notifications",
  "tool-workspacePerformanceReportTool": "Building workspace report",
  "tool-workspaceCommandCenterReportTool": "Building command center",
  "tool-pulseReportTool": "Building workspace pulse",
  "tool-storyPerformanceReportTool": "Building story report",
  "tool-objectiveProgressReportTool": "Building objective report",
  "tool-teamPerformanceReportTool": "Building team report",
  "tool-sprintPerformanceReportTool": "Building sprint report",
  "tool-timelineTrendsReportTool": "Building trends report",
  "tool-workloadPlanningTool": "Analyzing workload",
  "tool-focusBrief": "Reviewing priorities",
  "tool-mayaWorkPlanTool": "Planning work",
  "tool-getGitHubIntegrationTool": "Checking GitHub integration",
  "tool-createGitHubInstallSessionTool": "Creating GitHub install link",
  "tool-resyncGitHubRepositoriesTool": "Resyncing GitHub repositories",
  "tool-createGitHubIssueSyncLinkTool": "Linking GitHub repository",
  "tool-deleteGitHubIssueSyncLinkTool": "Removing GitHub sync link",
  "tool-updateGitHubWorkspaceSettingsTool": "Updating GitHub settings",
  "tool-getGitHubTeamSettingsTool": "Checking GitHub automation",
  "tool-updateGitHubTeamSettingsTool": "Updating GitHub automation",
  "tool-getStoryGitHubLinksTool": "Checking story GitHub links",
  "tool-getStoryGitHubCommentsTool": "Reading GitHub comments",
  "tool-postStoryGitHubCommentTool": "Posting GitHub comment",
  "tool-deleteStoryGitHubLinkTool": "Removing story GitHub link",
  "tool-links": "Loading links",
  "tool-labels": "Loading labels",
  "tool-storyLabels": "Managing labels",
  "tool-storyActivities": "Loading activity",
  "tool-listAttachments": "Loading attachments",
  "tool-deleteAttachment": "Deleting attachment",
  "tool-listMemories": "Checking memory",
  "tool-createMemory": "Saving to memory",
  "tool-updateMemory": "Updating memory",
  "tool-deleteMemory": "Removing memory",
  "tool-listCustomerFeedbackTool": "Reading customer feedback",
  "tool-getCustomerFeedbackTool": "Getting feedback details",
  "tool-theme": "Changing theme",
};

const DEFAULT_PROGRESS_LABEL = "Working on it";

export const getMessageProgressLabel = (message: MayaUIMessage) => {
  const lastPart = message.parts.at(-1);

  if (lastPart?.type === "text" && lastPart.text.trim()) {
    return undefined;
  }

  let latestToolPart: ToolMessagePart | undefined;
  for (const part of message.parts) {
    if (isToolMessagePart(part) && !isSupportingToolType(part.type)) {
      latestToolPart = part;
    }
  }

  if (!latestToolPart) {
    return "Thinking";
  }

  return TOOL_THINKING_LABELS[latestToolPart.type] ?? DEFAULT_PROGRESS_LABEL;
};

export const getMessageText = (message: MayaUIMessage) => {
  let text = "";
  for (const part of message.parts) {
    if (part.type === "text") {
      text += part.text;
    }
  }
  return text;
};

export const hasVisibleMessageContent = (message: MayaUIMessage) => {
  if (message.role === "user") {
    return true;
  }

  if (getMessageText(message).trim()) {
    return true;
  }

  if (message.parts.some((part) => part.type === "file")) {
    return true;
  }

  return message.parts.some(
    (part) => isToolMessagePart(part) && isRenderableToolPart(part),
  );
};
