export type JsonRecord = Record<string, unknown>;
export type Workspace = {
  id: string;
  name: string;
  slug: string;
  color?: string;
};
export type User = { id: string; fullName: string; email: string };
export type Bootstrap = {
  user: User | null;
  workspaces: Workspace[];
  activeWorkspaceId: string | null;
  needsSignIn: boolean;
};
export type Story = JsonRecord & {
  id: string;
  title: string;
  description?: string;
  descriptionHTML?: string;
  statusId: string;
  teamId: string;
  assigneeId?: string | null;
  sequenceId?: number;
  reference?: string;
  priority: string;
  updatedAt: string;
};
export type Document = JsonRecord & {
  id: string;
  title: string;
  contentHtml: string;
  contentText: string;
  revision: number;
  collaborationEpoch: number;
  collaborationState?: string;
  canEdit: boolean;
  updatedAt: string;
};
export type Team = { id: string; name: string; code: string };
export type Status = {
  id: string;
  name: string;
  color?: string;
  teamId: string;
  category?: string;
};
export type Comment = {
  id: string;
  storyId: string;
  content?: string;
  body?: string;
  createdAt: string;
  authorName?: string;
  comment?: string;
  subComments?: Comment[];
};
export type Operation = {
  id: string;
  kind: string;
  resourceId: string;
  payload: JsonRecord;
  baseVersion: string | null;
  status: "pending" | "conflict" | "blocked";
  error: string | null;
  serverVersion?: string;
  serverValue?: JsonRecord;
};
export type Attachment = {
  id: string;
  resourceId: string;
  targetKind: "story" | "document";
  filename: string;
  size: number;
  mimeType?: string;
};
export type WorkspaceData = {
  attachments?: Attachment[];
  notifications?: AppNotification[];
  unreadNotificationCount?: number;
  downloadState?: DownloadState;
  complete?: {
    stories: boolean;
    documents: boolean;
    comments: boolean;
    teams: boolean;
    statuses: boolean;
  };
  workspaceId: string;
  stories: Story[];
  documents: Document[];
  comments: Comment[];
  commentsCompleteFor?: string[];
  teams: Team[];
  statuses: Status[];
  pending: Operation[];
  lastSyncedAt: string | null;
};
export type Settings = {
  notifications: boolean;
  launchAtLogin: boolean;
  theme: "system" | "light" | "dark";
  appVersion: string;
  updateConfigured: boolean;
};
export type DraftKind = "story" | "document" | "comment" | "collaboration";
export type OperationInput = {
  consumedDraft?: {
    resourceId: string;
    kind: "story" | "document" | "comment";
  };
  workspaceId: string;
  kind: string;
  resourceId: string;
  payload: JsonRecord;
  baseVersion: string | null;
};

export type WorkspaceSettings = {
  profile: JsonRecord;
  workspace: JsonRecord;
  teams: JsonRecord[];
  members: JsonRecord[];
  preferences: JsonRecord;
  downloadedAt: string | null;
};

export type DownloadState = {
  workspaceId: string;
  stage: "my-work" | "settings" | "recent" | "documents" | "ready" | "paused";
  downloaded: number;
  error?: string;
};
export type AppNotification = {
  id: string;
  entityType: string;
  entityId: string;
  title: string;
  message?:
    | string
    | { template: string; variables: Record<string, { value: string }> };
  createdAt: string;
  readAt: string | null;
};
