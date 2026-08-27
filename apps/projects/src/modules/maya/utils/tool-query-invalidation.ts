import {
  analyticsKeys,
  calendarKeys,
  githubKeys,
  integrationRequestKeys,
  labelKeys,
  linkKeys,
  memberKeys,
  notificationKeys,
  sprintKeys,
  statusKeys,
  teamKeys,
} from "@/constants/keys";
import { isMutationToolCall, type MayaToolName } from "@/lib/ai/tool-policy";
import { aiChatKeys } from "@/modules/ai-chats/constants";
import { objectiveKeys } from "@/modules/objectives/constants";
import { storyKeys } from "@/modules/stories/constants";

type InvalidationProfile =
  | "github-integration"
  | "github-request"
  | "github-story"
  | "integration-request"
  | "labels"
  | "memories"
  | "none"
  | "notifications"
  | "objective-lifecycle"
  | "objective-statuses"
  | "sprint-settings"
  | "statuses"
  | "story-details"
  | "story-lifecycle"
  | "story-links"
  | "team-lifecycle"
  | "work-plan";

export const MAYA_TOOL_INVALIDATION_PROFILES = {
  acceptAllIntegrationRequestsTool: "integration-request",
  acceptIntegrationRequestTool: "integration-request",
  addStoryAssociation: "story-details",
  applyMayaWorkPlanTool: "work-plan",
  assignStoriesToUser: "story-lifecycle",
  bulkCreateStories: "story-lifecycle",
  bulkDeleteStories: "story-lifecycle",
  bulkUpdateStories: "story-lifecycle",
  comments: "story-details",
  createGitHubInstallSessionTool: "none",
  createGitHubIssueSyncLinkTool: "github-integration",
  createKeyResultTool: "objective-lifecycle",
  createMemory: "memories",
  createObjectiveTool: "objective-lifecycle",
  createStory: "story-lifecycle",
  createTeamTool: "team-lifecycle",
  declineAllIntegrationRequestsTool: "integration-request",
  declineIntegrationRequestTool: "integration-request",
  deleteAttachment: "story-details",
  deleteGitHubIssueSyncLinkTool: "github-integration",
  deleteKeyResultTool: "objective-lifecycle",
  deleteMemory: "memories",
  deleteObjectiveTool: "objective-lifecycle",
  deleteStory: "story-lifecycle",
  deleteStoryGitHubLinkTool: "github-story",
  deleteTeam: "team-lifecycle",
  duplicateStory: "story-lifecycle",
  joinTeam: "team-lifecycle",
  labels: "labels",
  leaveTeam: "team-lifecycle",
  links: "story-links",
  notifications: "notifications",
  objectiveStatuses: "objective-statuses",
  postRequestGitHubCommentTool: "github-request",
  postStoryGitHubCommentTool: "github-story",
  removeStoryAssociation: "story-details",
  resyncGitHubRepositoriesTool: "github-integration",
  restoreStory: "story-lifecycle",
  statuses: "statuses",
  storyLabels: "story-details",
  updateGitHubTeamSettingsTool: "github-integration",
  updateGitHubWorkspaceSettingsTool: "github-integration",
  updateIntegrationRequestTool: "integration-request",
  updateKeyResultTool: "objective-lifecycle",
  updateMemory: "memories",
  updateObjectiveTool: "objective-lifecycle",
  updateSprintSettings: "sprint-settings",
  updateStory: "story-lifecycle",
  updateTeam: "team-lifecycle",
} as const satisfies Partial<Record<MayaToolName, InvalidationProfile>>;

const asRecord = (value: unknown): Record<string, unknown> =>
  value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {};

const uniqueQueryKeys = (queryKeys: readonly (readonly unknown[])[]) => {
  const seen = new Set<string>();
  return queryKeys.filter((queryKey) => {
    const fingerprint = JSON.stringify(queryKey);
    if (seen.has(fingerprint)) return false;
    seen.add(fingerprint);
    return true;
  });
};

export const getMayaToolInvalidationKeys = ({
  input,
  toolName,
  workspaceSlug,
}: {
  input: unknown;
  toolName: string;
  workspaceSlug: string;
}): readonly (readonly unknown[])[] => {
  if (!isMutationToolCall(toolName, input)) return [];

  const profile =
    MAYA_TOOL_INVALIDATION_PROFILES[
      toolName as keyof typeof MAYA_TOOL_INVALIDATION_PROFILES
    ];
  const storyQueryKeys = [
    storyKeys.all(workspaceSlug),
    storyKeys.total(workspaceSlug),
  ] as const;
  const inputRecord = asRecord(input);

  let queryKeys: readonly (readonly unknown[])[];
  switch (profile) {
    case "story-lifecycle":
      queryKeys = [
        ...storyQueryKeys,
        calendarKeys.all(workspaceSlug),
        analyticsKeys.all(workspaceSlug),
      ];
      break;
    case "story-details":
      queryKeys = [storyKeys.all(workspaceSlug)];
      break;
    case "story-links":
      queryKeys = [
        storyKeys.all(workspaceSlug),
        typeof inputRecord.storyId === "string"
          ? linkKeys.story(inputRecord.storyId)
          : (["story-links"] as const),
      ];
      break;
    case "team-lifecycle":
      queryKeys = [
        teamKeys.all(workspaceSlug),
        memberKeys.all(workspaceSlug),
        ...storyQueryKeys,
        calendarKeys.all(workspaceSlug),
        analyticsKeys.all(workspaceSlug),
      ];
      break;
    case "objective-lifecycle":
      queryKeys = [
        objectiveKeys.all(workspaceSlug),
        storyKeys.all(workspaceSlug),
        analyticsKeys.all(workspaceSlug),
      ];
      break;
    case "sprint-settings":
      queryKeys = [
        sprintKeys.all(workspaceSlug),
        teamKeys.all(workspaceSlug),
        storyKeys.all(workspaceSlug),
        calendarKeys.all(workspaceSlug),
        analyticsKeys.all(workspaceSlug),
      ];
      break;
    case "work-plan":
      queryKeys = [
        calendarKeys.all(workspaceSlug),
        storyKeys.all(workspaceSlug),
        analyticsKeys.all(workspaceSlug),
      ];
      break;
    case "github-integration":
      queryKeys = [githubKeys.all(workspaceSlug)];
      break;
    case "github-story":
      queryKeys = [githubKeys.all(workspaceSlug), storyKeys.all(workspaceSlug)];
      break;
    case "integration-request":
      queryKeys = [integrationRequestKeys.all(workspaceSlug)];
      break;
    case "github-request":
      queryKeys = [
        integrationRequestKeys.all(workspaceSlug),
        githubKeys.all(workspaceSlug),
      ];
      break;
    case "labels":
      queryKeys = [labelKeys.all(workspaceSlug), storyKeys.all(workspaceSlug)];
      break;
    case "notifications":
      queryKeys = [notificationKeys.all(workspaceSlug)];
      break;
    case "objective-statuses":
      queryKeys = [
        objectiveKeys.all(workspaceSlug),
        analyticsKeys.all(workspaceSlug),
      ];
      break;
    case "statuses":
      queryKeys = [
        statusKeys.all(workspaceSlug),
        storyKeys.all(workspaceSlug),
        analyticsKeys.all(workspaceSlug),
      ];
      break;
    case "memories":
      queryKeys = [aiChatKeys.memories()];
      break;
    case "none":
    default:
      queryKeys = [];
      break;
  }

  return uniqueQueryKeys(queryKeys);
};
