export type SlackWorkspaceSettings = {
  defaultCreateMode: "create_task_now" | "send_to_requests";
  createdAt: string;
  updatedAt: string;
};

export type SlackWorkspace = {
  id: string;
  slackTeamId: string;
  slackTeamName: string;
  slackTeamDomain: string;
  botUserId?: string | null;
  scope?: string | null;
  isActive: boolean;
  installedByUserId?: string | null;
  createdAt: string;
  updatedAt: string;
};

export type SlackChannel = {
  id: string;
  slackChannelId: string;
  name: string;
  isPrivate: boolean;
  isArchived: boolean;
  isMember: boolean;
  isActive: boolean;
  lastSyncedAt?: string | null;
  createdAt: string;
  updatedAt: string;
};

export type SlackChannelLink = {
  id: string;
  slackChannelId: string;
  teamId: string;
  teamCode: string;
  teamName: string;
  teamColor: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};

export type SlackIntegration = {
  settings: SlackWorkspaceSettings;
  slackWorkspace?: SlackWorkspace | null;
  channels: SlackChannel[];
  channelLinks: SlackChannelLink[];
};

export type CreateSlackInstallSessionResponse = {
  installUrl: string;
};

export type UpdateSlackWorkspaceSettingsInput = Partial<{
  defaultCreateMode: "create_task_now" | "send_to_requests";
}>;

export type CreateSlackChannelLinkInput = {
  slackChannelId: string;
  teamId: string;
};
