export type DisplayColumn =
  | "ID"
  | "Status"
  | "Assignee"
  | "Estimate"
  | "Priority"
  | "Deadline"
  | "Created"
  | "Updated"
  | "Sprint"
  | "Objective"
  | "Key Result"
  | "Epic"
  | "Labels";

export const DISPLAY_COLUMNS_VERSION = 2;

export const migrateDisplayColumns = (
  displayColumns: DisplayColumn[],
  version = 1,
) => {
  if (
    version >= DISPLAY_COLUMNS_VERSION ||
    !displayColumns.includes("Objective") ||
    displayColumns.includes("Key Result")
  ) {
    return displayColumns;
  }

  const migratedColumns = [...displayColumns];
  const objectiveIndex = migratedColumns.indexOf("Objective");
  migratedColumns.splice(objectiveIndex + 1, 0, "Key Result");
  return migratedColumns;
};
