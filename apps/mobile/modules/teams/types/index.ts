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
  sprintsEnabled: boolean;
};

export type TeamStoriesTab = "all" | "active" | "backlog";

export type TeamViewOptions = {
  groupBy: "status" | "priority" | "assignee";
  orderBy: "created" | "updated" | "deadline" | "priority";
  orderDirection: "asc" | "desc";
};
