export const VIEW_ICON_KEYS = [
  "list",
  "kanban",
  "calendar",
  "clock",
  "star",
  "analytics",
  "goal",
  "tags",
  "code",
  "team",
  "book",
  "workflow",
  "roadmap",
  "docs",
] as const;

export type ViewIconKey = (typeof VIEW_ICON_KEYS)[number];
