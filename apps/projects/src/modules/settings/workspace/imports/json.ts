import type { ImportAnalysis, ImportDraft } from "./schema";
import { IMPORT_MAX_TASKS } from "./schema";
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
  const records = findTaskRecords(value);
  if (records.length === 0) {
    throw new Error(
      "The JSON file must contain a task array or a tasks, items, issues, cards, records, or data collection.",
    );
  }

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
    summary: `Found ${tasks.length} importable tasks across ${columns.length} JSON fields. Semantic mapping is prepared from the complete JSON document when AI analysis is available.`,
    warnings,
    mapping,
    tasks,
  };

  return { ...analysis, columns, fileHash, fileName, rows };
};
