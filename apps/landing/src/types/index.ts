export type UserRole = "admin" | "member" | "guest";

export type User = {
  id: string;
  username: string;
  email: string;
  fullName: string;
  avatarUrl: string;
  isActive: boolean;
  lastUsedWorkspaceId: string;
  createdAt: string;
  updatedAt: string;
};

export type Workspace = {
  id: string;
  name: string;
  slug: string;
  color: string;
  userRole: UserRole;
  createdAt: string;
  updatedAt: string;
};

export type ApiResponse<T> = {
  data: T | null;
  error?: {
    message: string;
  };
};

export type Invitation = {
  id: string;
  workspaceId: string;
  inviterId: string;
  email: string;
  role: string;
  teamIds: string[];
  expiresAt: string;
  token?: string;
  usedAt?: string;
  createdAt: string;
  updatedAt: string;
  workspaceName: string;
  workspaceSlug: string;
  workspaceColor: string;
};

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
