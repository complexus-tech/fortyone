export type Team = {
  id: string;
  name: string;
  code: string;
  color: string;
  isPrivate: boolean;
  workspaceId: string;
  createdAt: string;
  updatedAt: string;
  memberCount: number;
};

export type CreateTeamInput = {
  name: string;
  code: string;
  color: string;
  isPrivate: boolean;
};

export type UpdateTeamInput = Partial<CreateTeamInput>;

export type TeamSprintSettings = {
  sprintsEnabled: boolean;
  autoCreateSprints: boolean;
  upcomingSprintsCount: number;
  sprintDurationWeeks: number;
  sprintStartDay: string;
  moveIncompleteStoriesEnabled: boolean;
  createdAt: string;
  updatedAt: string;
};

export type TeamStoryAutomationSettings = {
  autoCloseInactiveEnabled: boolean;
  autoCloseInactiveMonths: number;
  autoArchiveEnabled: boolean;
  autoArchiveMonths: number;
  createdAt: string;
  updatedAt: string;
};

export type TeamSettings = {
  sprintSettings: TeamSprintSettings;
  storyAutomationSettings: TeamStoryAutomationSettings;
};

export type UpdateSprintSettingsInput = Partial<
  Omit<TeamSprintSettings, "createdAt" | "updatedAt">
>;

export type UpdateStoryAutomationSettingsInput = Partial<
  Omit<TeamStoryAutomationSettings, "createdAt" | "updatedAt">
>;

export type ReorderTeamsInput = {
  teamIds: string[];
};
