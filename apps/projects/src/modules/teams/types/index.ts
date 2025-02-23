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

export type CreateTeamInput = {
  name: string;
  code: string;
  color: string;
  isPrivate: boolean;
};

export type UpdateTeamInput = Partial<CreateTeamInput>;
