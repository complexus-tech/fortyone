export type TeamStoriesTab = "all" | "active" | "backlog";

export type TeamViewOptions = {
  groupBy: "status" | "priority" | "assignee";
  orderBy: "created" | "updated" | "deadline" | "priority";
  orderDirection: "asc" | "desc";
};
