export type DisplayColumn =
  | "ID"
  | "Status"
  | "Assignee"
  | "Estimate"
  | "Time needed"
  | "Priority"
  | "Deadline"
  | "Created"
  | "Updated"
  | "Sprint"
  | "Objective"
  | "Key Result"
  | "Epic"
  | "Labels";

export const DISPLAY_COLUMNS_VERSION = 3;

export const migrateDisplayColumns = (
  displayColumns: DisplayColumn[],
  version = 1,
) => {
  if (version >= DISPLAY_COLUMNS_VERSION) return displayColumns;

  const migratedColumns = [...displayColumns];

  if (
    version < 2 &&
    migratedColumns.includes("Objective") &&
    !migratedColumns.includes("Key Result")
  ) {
    const objectiveIndex = migratedColumns.indexOf("Objective");
    migratedColumns.splice(objectiveIndex + 1, 0, "Key Result");
  }

  if (
    version < 3 &&
    migratedColumns.includes("Estimate") &&
    !migratedColumns.includes("Time needed")
  ) {
    const estimateIndex = migratedColumns.indexOf("Estimate");
    migratedColumns.splice(estimateIndex + 1, 0, "Time needed");
  }

  return migratedColumns;
};
