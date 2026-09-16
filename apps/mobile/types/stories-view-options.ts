export type DisplayColumn = "ID" | "Status" | "Assignee" | "Priority";

export const DEFAULT_STORY_DISPLAY_COLUMNS: DisplayColumn[] = [
  "Status",
  "Assignee",
  "Priority",
];

export type StoriesViewOptions = {
  groupBy: "status" | "priority" | "assignee";
  orderBy: "created" | "updated" | "deadline" | "priority";
  orderDirection: "asc" | "desc";
  displayColumns: DisplayColumn[];
};
