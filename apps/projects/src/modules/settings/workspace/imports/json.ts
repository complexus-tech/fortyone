import type { ImportAnalysis, ImportDraft } from "./schema";
import { createEmptyImportEntityCollections, IMPORT_MAX_TASKS } from "./schema";
import { inferImportMapping, mapRowsToImportTasks } from "./csv";

const MAX_COLUMNS = 75;
const MAX_CELL_CHARACTERS = 20_000;
const GENERIC_COLLECTION_KEYS = [
  "tasks",
  "items",
  "issues",
  "cards",
  "records",
  "data",
] as const;
const TRELLO_BOARD_ID_KEYS = ["boardId", "_id", "id"] as const;

type JsonRecord = Record<string, unknown>;

const isRecord = (value: unknown): value is JsonRecord =>
  typeof value === "object" && value !== null && !Array.isArray(value);

const toCellValue = (value: unknown) => {
  if (value === null || value === undefined) return "";
  if (typeof value === "string") return value.slice(0, MAX_CELL_CHARACTERS);
  if (typeof value === "number" || typeof value === "boolean") {
    return String(value);
  }
  try {
    return JSON.stringify(value).slice(0, MAX_CELL_CHARACTERS);
  } catch {
    return "";
  }
};

const parseJson = (text: string): unknown => {
  try {
    return JSON.parse(text.replace(/^\uFEFF/, "")) as unknown;
  } catch {
    throw new Error("The JSON file is not valid JSON.");
  }
};

const findTaskRecords = (value: unknown): JsonRecord[] => {
  if (Array.isArray(value)) return value.filter(isRecord);
  if (!isRecord(value)) return [];

  for (const key of GENERIC_COLLECTION_KEYS) {
    const candidate = value[key];
    if (Array.isArray(candidate)) return candidate.filter(isRecord);
  }
  return [];
};

const stableTextHash = (value: string) => {
  let first = 0x811c9dc5;
  let second = 0x9e3779b9;
  for (let index = 0; index < value.length; index += 1) {
    const code = value.charCodeAt(index);
    first = (first * 33 + code) % 0x1_0000_0000;
    second = (second * 65_599 + code) % 0x1_0000_0000;
  }
  return `${first.toString(16).padStart(8, "0")}${second.toString(16).padStart(8, "0")}`;
};

const sanitizeContainerId = (value: unknown, prefix: string) => {
  let explicitId = "";
  if (typeof value === "string") {
    explicitId = value.normalize("NFKC").trim();
  } else if (typeof value === "number" && Number.isSafeInteger(value)) {
    explicitId = String(value);
  }
  if (!explicitId) return null;

  try {
    const encodedId = encodeURIComponent(explicitId);
    const namespace = `${prefix}${encodedId}`;
    if (namespace.length <= 300) return namespace;

    const suffix = `~${stableTextHash(encodedId)}`;
    const availableIdLength = 300 - prefix.length - suffix.length;
    const truncatedId = encodedId
      .slice(0, availableIdLength)
      .replace(/%(?:[0-9A-F])?$/u, "");
    return `${prefix}${truncatedId}${suffix}`;
  } catch {
    return null;
  }
};

const findSourceNamespace = (value: unknown) => {
  if (!isRecord(value)) return null;
  if (!Array.isArray(value.cards) || !Array.isArray(value.lists)) return null;
  const listIds = new Set(
    value.lists.filter(isRecord).flatMap((list) => {
      const id = list.id;
      return typeof id === "string" || typeof id === "number"
        ? [String(id)]
        : [];
    }),
  );
  const hasTrelloListRelationship = value.cards
    .filter(isRecord)
    .some((card) => {
      const listId = card.idList;
      return (
        (typeof listId === "string" || typeof listId === "number") &&
        listIds.has(String(listId))
      );
    });
  const hasTrelloBoardMarker =
    isRecord(value.prefs) || "idOrganization" in value || "labelNames" in value;
  if (!hasTrelloListRelationship || !hasTrelloBoardMarker) return null;

  for (const key of TRELLO_BOARD_ID_KEYS) {
    const sourceNamespace = sanitizeContainerId(value[key], "trello:board:");
    if (sourceNamespace) return sourceNamespace;
  }
  return null;
};

export const createJsonImportDraft = ({
  fileHash,
  fileName,
  text,
}: {
  fileHash: string;
  fileName: string;
  text: string;
}): ImportDraft => {
  const value = parseJson(text);
  if (!Array.isArray(value) && !isRecord(value)) {
    throw new Error("The JSON file must contain an object or array.");
  }
  const records = findTaskRecords(value);
  const sourceRecords = records.slice(0, IMPORT_MAX_TASKS);
  const columns = Array.from(
    sourceRecords.reduce<Set<string>>((result, record) => {
      for (const key of Object.keys(record)) {
        if (result.size >= MAX_COLUMNS) break;
        result.add(key);
      }
      return result;
    }, new Set()),
  );
  const rows = sourceRecords.map((record) =>
    Object.fromEntries(
      columns.map((column) => [column, toCellValue(record[column])]),
    ),
  );
  const mapping = inferImportMapping(columns);
  const tasks = mapRowsToImportTasks(rows, mapping);
  const warnings = [
    ...(records.length === 0
      ? [
          "No standard task collection was found for the initial preview. AI analysis can still map supported objects from the complete JSON document.",
        ]
      : []),
    ...(records.length > IMPORT_MAX_TASKS
      ? [
          `Only the first ${IMPORT_MAX_TASKS} records are included in this import.`,
        ]
      : []),
    ...(tasks.length < rows.length
      ? [
          `${rows.length - tasks.length} records have no mapped title and will be skipped unless you change the mapping.`,
        ]
      : []),
  ];

  const analysis: ImportAnalysis = {
    sourceType: "json",
    sourceNamespace: findSourceNamespace(value),
    summary:
      records.length > 0
        ? `Found ${tasks.length} importable tasks across ${columns.length} JSON fields. Semantic mapping is prepared from the complete JSON document when AI analysis is available.`
        : "No standard task collection was found. Semantic mapping will inspect the complete JSON document for teams, people, strategic pillars, objectives, key results, sprints, labels, and work items.",
    warnings,
    mapping,
    ...createEmptyImportEntityCollections(),
    tasks,
  };

  return { ...analysis, columns, fileHash, fileName, rows };
};
