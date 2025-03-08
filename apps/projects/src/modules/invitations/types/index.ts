export type NewInvitation = {
  email: string;
  role: string;
  teamIds?: string[];
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
