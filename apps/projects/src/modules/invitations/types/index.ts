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
  expiresAt: string;
  token?: string;
  createdAt: string;
  updatedAt: string;
};
