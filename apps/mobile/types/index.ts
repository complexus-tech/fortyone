export type ApiResponse<T> = {
  data?: T | null;
  error?: {
    message: string;
  };
};

export type UserRole = "admin" | "member" | "guest" | "system";

export type User = {
  id: string;
  username: string;
  email: string;
  fullName: string;
  avatarUrl: string | null;
  isActive: boolean;
  lastUsedWorkspaceId: string;
  hasSeenWalkthrough: boolean;
  timezone: string;
  createdAt: string;
  updatedAt: string;
};
