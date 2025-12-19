import { listAttachments, deleteAttachment } from "@/lib/ai/tools/attachments";
import {
  createObjectiveTool,
  deleteObjectiveTool,
  listObjectivesTool,
  listTeamObjectivesTool,
  updateObjectiveTool,
  getObjectiveDetailsTool,
  objectiveAnalyticsTool,
  getObjectiveActivitiesTool,
} from "@/lib/ai/tools/objectives";
import {
  listKeyResultsTool,
  createKeyResultTool,
  updateKeyResultTool,
  deleteKeyResultTool,
  getKeyResultActivitiesTool,
} from "@/lib/ai/tools/key-results";
import {
  listTeamStories,
  searchStories,
  getStoryDetails,
  createStory,
  updateStory,
  deleteStory,
  bulkUpdateStories,
  bulkDeleteStories,
  bulkCreateStories,
  assignStoriesToUser,
  duplicateStory,
  restoreStory,
  addStoryAssociation,
  removeStoryAssociation,
} from "@/lib/ai/tools/stories";
import {
  listSprints,
  listRunningSprints,
  getSprintDetailsTool,
  getSprintAnalyticsTool,
  updateSprintSettings,
} from "@/lib/ai/tools/sprints";
import {
  listTeams,
  listPublicTeams,
  getTeamDetails,
  listTeamMembers,
  createTeamTool,
  updateTeam,
  joinTeam,
  deleteTeam,
  leaveTeam,
  getTeamSettingsTool,
} from "@/lib/ai/tools/teams";
import { navigation } from "./navigation";
import { membersTool } from "./members";
import { theme } from "./theme";
import { searchTool } from "./search";
import { linksTool } from "./links";
import { labelsTool } from "./labels";
import { storyActivitiesTool } from "./story-activities";
import { storyLabelsTool } from "./story-labels";
import { objectiveStatusesTool } from "./objective-statuses";
import { statusesTool } from "./statuses";
import { commentsTool } from "./comments";
import { notificationsTool } from "./notifications";
import { suggestions } from "./suggestions";

export { navigation } from "./navigation";
export { membersTool } from "./members";
export { statusesTool } from "./statuses";
export { objectiveStatusesTool } from "./objective-statuses";
export { searchTool } from "./search";
export { notificationsTool } from "./notifications";
export { theme } from "./theme";
export { commentsTool } from "./comments";
export { storyActivitiesTool } from "./story-activities";
export { linksTool } from "./links";
export { labelsTool } from "./labels";
export { storyLabelsTool } from "./story-labels";

export const tools = {
  navigation,
  theme,
  suggestions,
  members: membersTool,
  search: searchTool,
  notifications: notificationsTool,
  comments: commentsTool,
  // Teams
  listTeams,
  listPublicTeams,
  getTeamDetails,
  listTeamMembers,
  createTeamTool,
  updateTeam,
  joinTeam,
  deleteTeam,
  leaveTeam,
  getTeamSettingsTool,
  // Stories
  listTeamStories,
  searchStories,
  getStoryDetails,
  createStory,
  updateStory,
  deleteStory,
  bulkUpdateStories,
  bulkDeleteStories,
  bulkCreateStories,
  assignStoriesToUser,
  duplicateStory,
  restoreStory,
  addStoryAssociation,
  removeStoryAssociation,
  statuses: statusesTool,
  // Sprints
  listSprints,
  listRunningSprints,
  getSprintDetailsTool,
  getSprintAnalyticsTool,
  updateSprintSettings,
  objectiveStatuses: objectiveStatusesTool,
  // Key Results
  listKeyResultsTool,
  createKeyResultTool,
  updateKeyResultTool,
  deleteKeyResultTool,
  getKeyResultActivitiesTool,
  // Attachments
  listAttachments,
  deleteAttachment,
  // Objectives
  listObjectivesTool,
  listTeamObjectivesTool,
  createObjectiveTool,
  updateObjectiveTool,
  deleteObjectiveTool,
  objectiveAnalyticsTool,
  getObjectiveDetailsTool,
  getObjectiveActivitiesTool,
  // Links
  links: linksTool,
  labels: labelsTool,
  storyActivities: storyActivitiesTool,
  storyLabels: storyLabelsTool,
};
