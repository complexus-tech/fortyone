export type Pagination = {
  total: number;
  page: number;
  limit: number;
  offset: number;
};

export type ListResult<T> = {
  items: T[];
  pagination: Pagination;
};

export type DashboardSummary = {
  totalWorkspaces: number;
  activeTrials: number;
  expiredTrials: number;
  paidWorkspaces: number;
  deletedWorkspaces: number;
  totalUsers: number;
  internalUsers: number;
  activeSubscriptions: number;
  slackInstallations: number;
  gitHubInstallations: number;
  recentAdminAuditLogs: number;
};

export type UserSummary = {
  id: string;
  username: string;
  email: string;
  fullName: string;
  avatarUrl: string;
  isActive: boolean;
  isSystem: boolean;
  isInternal: boolean;
  lastLoginAt: string | null;
  lastUsedWorkspaceId: string | null;
  lastUsedWorkspace: string | null;
  gitHubUsername: string | null;
  workspaceCount: number;
  createdAt: string;
  updatedAt: string;
};

export type UserMembership = {
  workspaceId: string;
  workspaceName: string;
  workspaceSlug: string;
  role: string;
  joinedAt: string;
};

export type UserOverview = {
  user: UserSummary;
  memberships: UserMembership[];
};

export type WorkspaceSummary = {
  id: string;
  name: string;
  slug: string;
  avatarUrl: string | null;
  color: string;
  teamSize: string;
  trialEndsOn: string | null;
  deletedAt: string | null;
  lastAccessedAt: string | null;
  createdByUserId: string | null;
  createdByEmail: string | null;
  createdByName: string | null;
  memberCount: number;
  teamCount: number;
  storyCount: number;
  subscriptionTier: string | null;
  subscriptionStatus: string | null;
  subscriptionSeats: number | null;
  stripeCustomerId: string | null;
  stripeSubscriptionId: string | null;
  slackInstalled: boolean;
  gitHubInstalled: boolean;
  createdAt: string;
  updatedAt: string;
};

export type WorkspaceMember = {
  userId: string;
  email: string;
  fullName: string;
  role: string;
  isInternal: boolean;
  joinedAt: string;
};

export type WorkspaceOverview = {
  workspace: WorkspaceSummary;
  members: WorkspaceMember[];
};

export type AuditLog = {
  id: string;
  actorUserId: string;
  actorEmail: string;
  actorName: string;
  targetType: string;
  targetId: string | null;
  workspaceId: string | null;
  workspaceName: string | null;
  workspaceSlug: string | null;
  action: string;
  fieldName: string;
  oldValue: unknown;
  newValue: unknown;
  reason: string;
  metadata: unknown;
  createdAt: string;
};

export type AdminListParams = {
  q?: string;
  status?: string;
  page?: string | number;
  limit?: string | number;
};
