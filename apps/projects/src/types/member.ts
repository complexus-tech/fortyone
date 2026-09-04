import type { UserRole } from "./user-role";

export type Member = {
  id: string;
  username: string;
  email: string;
  role: UserRole;
  fullName: string;
  avatarUrl: string | null;
  isActive: boolean;
  isSystem: boolean;
  isInternal: boolean;
  teamAiRoleTitle?: string;
  teamAiRoleDescription?: string;
  inferredTeamAiRoleTitle?: string;
  inferredTeamAiRoleDescription?: string;
  inferredTeamAiRoleStoryCount?: number;
  inferredTeamAiRoleConfidence?: number;
  inferredTeamAiRoleGeneratedAt?: string | null;
  lastStoryActivityAt?: string | null;
  createdAt: string;
  updatedAt: string;
};

export type MembersPage = {
  members: Member[];
  pagination: {
    page: number;
    pageSize: number;
    hasMore: boolean;
    nextPage: number;
  };
};
