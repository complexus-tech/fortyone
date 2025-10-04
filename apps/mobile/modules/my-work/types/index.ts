export type MyWorkTab = "all" | "assigned" | "created";

export type MyWorkViewOptions = {
  groupBy: "status" | "priority" | "assignee";
  orderBy: "created" | "updated" | "deadline" | "priority";
  orderDirection: "asc" | "desc";
};
