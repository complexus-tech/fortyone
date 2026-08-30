export type User = {
  id: string;
  username: string;
  email: string;
  fullName: string;
  avatarUrl: string | null;
  isActive: boolean;
  isSystem: boolean;
  isInternal: boolean;
  lastUsedWorkspaceId: string;
  hasSeenWalkthrough: boolean;
  timezone: string;
  workingDays?: number[] | null;
  workingStartMinute?: number | null;
  workingEndMinute?: number | null;
  githubUsername: string | null;
  createdAt: string;
  updatedAt: string;
};
