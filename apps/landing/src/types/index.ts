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
