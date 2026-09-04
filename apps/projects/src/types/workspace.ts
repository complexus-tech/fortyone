import type { UserRole } from "./user-role";

export type Workspace = {
  id: string;
  name: string;
  slug: string;
  color: string;
  avatarUrl: string | null;
  userRole: UserRole;
  trialEndsOn: string | null;
  deletedAt: string | null;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};
