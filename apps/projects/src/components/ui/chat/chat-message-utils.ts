import type { MayaUIMessage } from "@/lib/ai/tools/types";
import type { ToolMessagePart } from "./tool-output-policy";
import {
  isRenderableToolPart,
  isSupportingToolType,
  isToolMessagePart,
} from "./tool-output-policy";

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
  "tool-comments": "Loading comments",
  "tool-notifications": "Checking notifications",
  "tool-workspacePerformanceReportTool": "Building workspace report",
  "tool-workspaceCommandCenterReportTool": "Building command center",
  "tool-storyPerformanceReportTool": "Building story report",
  "tool-objectiveProgressReportTool": "Building objective report",
  "tool-teamPerformanceReportTool": "Building team report",
  "tool-sprintPerformanceReportTool": "Building sprint report",
  "tool-timelineTrendsReportTool": "Building trends report",
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
