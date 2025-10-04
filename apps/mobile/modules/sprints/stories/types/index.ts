export type SprintStoriesTab = "all" | "active" | "completed";

export type SprintViewOptions = {
  groupBy: "status" | "priority" | "assignee";
  orderBy: "created" | "updated" | "deadline" | "priority";
  orderDirection: "asc" | "desc";
};
