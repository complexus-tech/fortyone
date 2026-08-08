export type MyWorkLayout = "list" | "kanban";

export const normalizeMyWorkLayout = (
  value: unknown,
  fallback: MyWorkLayout,
): MyWorkLayout => (value === "list" || value === "kanban" ? value : fallback);
