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

export type SlackIntegration = {
  slackWorkspace?: SlackWorkspace | null;
  channels: SlackChannel[];
};

export type CreateSlackInstallSessionResponse = {
  installUrl: string;
};

export type SlackRequestLog = {
  id: string;
  requestType: string;
  endpoint: string;
  workspaceId?: string | null;
  slackTeamId?: string | null;
  slackUserId?: string | null;
  slackChannelId?: string | null;
  command?: string | null;
  triggerId?: string | null;
  requestBody?: string | null;
  headers: Record<string, string>;
  responseCode: number;
  outcome: string;
  errorMessage?: string | null;
  createdAt: string;
};
