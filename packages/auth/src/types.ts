export type CurrentUser = {
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

export type ApiResponse<T> = {
  data?: T | null;
  error?: {
    message: string;
  };
};
