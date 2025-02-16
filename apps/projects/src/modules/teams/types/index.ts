export type TeamMember = {
  id: string;
  userId: string;
  teamId: string;
  role: "admin" | "member";
  joinedAt: string;
};

export type Team = {
  id: string;
  name: string;
  code: string;
  color: string;
  workspaceId: string;
  createdAt: string;
  updatedAt: string;
  members: TeamMember[];
};

export type CreateTeamInput = {
  name: string;
  code: string;
  color: string;
};

export type UpdateTeamInput = Partial<CreateTeamInput>;
