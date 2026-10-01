import type {
  ImportDraft,
  ImportPerson,
  ImportTask,
  ImportLabel,
} from "./schema";
import { IMPORT_MAX_TASKS } from "./schema";

const normalize = (value: string) =>
  value.trim().toLowerCase().replace(/[_-]+/g, " ").replace(/\s+/g, " ");
const splitList = (value: string) =>
  value
    .split(/[;\n]/)
    .map((item) => item.trim())
    .filter(Boolean);

// CSV exports describe values by column name. Do not infer provider accounts,
// money units, private membership, or typed fields from incidental text.
export const enrichVendorExport = (draft: ImportDraft): ImportDraft => {
  const columns = new Map(
    draft.columns.map((column) => [normalize(column), column]),
  );
  const shortcut =
    columns.has("workflow id") &&
    columns.has("is completed") &&
    columns.has("name");
  const plane =
    columns.has("sequence id") && columns.has("project") && columns.has("name");
  if (!shortcut && !plane) return draft;
  const platform = shortcut ? "Shortcut" : "Plane";
  const cell = (row: Record<string, string>, name: string) => {
    const column = columns.get(name);
    return column && Object.prototype.hasOwnProperty.call(row, column)
      ? row[column].trim()
      : "";
  };
  const people = new Map<string, ImportPerson>();
  const labels = new Map<string, ImportLabel>();
  const taskById = new Map(draft.tasks.map((task) => [task.sourceId, task]));
  if (taskById.size !== draft.tasks.length)
    return {
      ...draft,
      warnings: [
        ...draft.warnings,
        `${platform} source IDs are duplicated. All rows remain available through the reviewed generic mapping.`,
      ],
    };
  const tasks: ImportTask[] = [];
  let mappedRows = 0;
  let opaqueCustomFields = 0;
  for (const row of draft.rows) {
    const sourceId = cell(row, "id");
    const mapped = taskById.get(sourceId);
    if (!mapped) continue;
    mappedRows += 1;
    const labelIds = splitList(cell(row, "labels")).map((name) => {
      const labelId = `${platform.toLowerCase()}:label:${name}`.slice(0, 300);
      labels.set(labelId, {
        sourceId: labelId,
        name: name.slice(0, 100),
        color: null,
        teamSourceId: null,
      });
      return labelId;
    });
    const owners = splitList(cell(row, shortcut ? "owners" : "assignees"));
    const ownerIds = owners.map((owner) => {
      const id = `${platform.toLowerCase()}:person:${owner}`.slice(0, 300);
      people.set(id, {
        sourceId: id,
        name: owner.includes("@") ? null : owner.slice(0, 255),
        email: owner.includes("@") ? owner : null,
        teamSourceIds: [],
      });
      return id;
    });
    if (cell(row, "custom fields")) opaqueCustomFields += 1;
    const task: ImportTask = {
      ...mapped,
      labelSourceIds: labelIds.slice(0, 100),
      assigneePersonSourceId: ownerIds[0] ?? null,
      collaboratorPersonSourceIds: ownerIds.slice(1, 101),
      assigneeName: owners[0] ?? mapped.assigneeName,
      status: cell(row, "state") || mapped.status,
      statusCategory:
        cell(row, "is completed").toLowerCase() === "true"
          ? "completed"
          : mapped.statusCategory,
    };
    tasks.push(task);
    if (shortcut) {
      splitList(cell(row, "tasks")).forEach((value, index) => {
        if (!/^\[[xX ]\]\s*\S/.test(value)) return;
        const checked = value.charAt(1);
        const title = value.replace(/^\[[xX ]\]\s*/, "");
        tasks.push({
          ...task,
          sourceId: `shortcut:task:${sourceId}:${index + 1}`,
          title: title.slice(0, 255),
          description: "Shortcut checklist item",
          parentSourceId: task.sourceId,
          status: checked.toLowerCase() === "x" ? "Completed" : "To do",
          statusCategory:
            checked.toLowerCase() === "x" ? "completed" : "unstarted",
          labelSourceIds: [],
          collaboratorPersonSourceIds: [],
          assigneePersonSourceId: null,
          assigneeName: null,
        });
      });
    }
  }
  if (mappedRows !== draft.tasks.length)
    return {
      ...draft,
      warnings: [
        ...draft.warnings,
        `${platform} columns were recognized, but stable source IDs could not be mapped. Review the ID mapping.`,
      ],
    };
  return {
    ...draft,
    mapping: null,
    tasks: tasks.slice(0, IMPORT_MAX_TASKS),
    people: [...people.values()].slice(0, 500),
    labels: [...labels.values()].slice(0, 500),
    sourceMetadata: {
      platform: shortcut ? "shortcut" : "plane",
      archivedTaskSourceIds: [],
      nestedChecklistItemCount: tasks.length - draft.tasks.length,
    },
    summary: `Recognized ${platform} export columns. Review ${tasks.length} work items, workflow states, labels and people before import.`,
    warnings: [
      ...draft.warnings,
      `${platform} files do not establish an authenticated provider connection. Comments or file contents absent from the export cannot be recovered.`,
      "The same file can be retried safely. For a refreshed export, preserve a source workspace identifier before importing to avoid creating a separate source set.",
      ...(opaqueCustomFields
        ? [
            `${opaqueCustomFields} rows include custom field data without typed definitions. Keep the export as your source record; those fields require explicit type mapping.`,
          ]
        : []),
      ...(tasks.length > IMPORT_MAX_TASKS
        ? [
            `Split the export: only ${IMPORT_MAX_TASKS} work items fit one review graph.`,
          ]
        : []),
    ].slice(0, 50),
  };
};
