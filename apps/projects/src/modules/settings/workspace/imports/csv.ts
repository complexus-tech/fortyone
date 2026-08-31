import { format, isValid, parse } from "date-fns";
import type {
  ImportAnalysis,
  ImportDraft,
  ImportMapping,
  ImportPriority,
  ImportSourceType,
  ImportTask,
} from "./schema";
import { IMPORT_MAX_TASKS } from "./schema";

const MAX_COLUMNS = 75;
const MAX_CELL_CHARACTERS = 20_000;

const FIELD_ALIASES: Record<keyof ImportMapping, readonly string[]> = {
  title: [
    "summary",
    "issue summary",
    "title",
    "task",
    "task name",
    "story",
    "story title",
    "subject",
    "name",
  ],
  description: [
    "description",
    "desc",
    "details",
    "body",
    "notes",
    "task details",
  ],
  status: ["status", "state", "workflow status", "issue status"],
  priority: ["priority", "urgency", "issue priority"],
  assigneeEmail: [
    "assignee email",
    "assignee",
    "assigned to",
    "owner email",
    "owner",
  ],
  startDate: ["start date", "start", "planned start"],
  endDate: ["due date", "due", "deadline", "end date", "end", "target date"],
  sourceId: ["issue key", "key", "issue id", "task id", "story id", "id"],
};

const DATE_PATTERNS = [
  "yyyy-MM-dd",
  "yyyy/MM/dd",
  "M/d/yyyy",
  "MM/dd/yyyy",
  "d/M/yyyy",
  "dd/MM/yyyy",
  "d/MMM/yy h:mm a",
  "dd/MMM/yy h:mm a",
  "d/MMM/yyyy h:mm a",
  "dd/MMM/yyyy h:mm a",
  "d/MMM/yy H:mm",
  "dd/MMM/yy H:mm",
] as const;
const NUMERIC_DATE_PART_PATTERN = /^\d{1,2}$/;
const NUMERIC_DATE_YEAR_PATTERN = /^\d{4}$/;

const isAmbiguousNumericDate = (value: string) => {
  const parts = value.trim().split("/");
  const [first, second, year] = parts;
  if (
    parts.length !== 3 ||
    !first ||
    !second ||
    !year ||
    !NUMERIC_DATE_PART_PATTERN.test(first) ||
    !NUMERIC_DATE_PART_PATTERN.test(second) ||
    !NUMERIC_DATE_YEAR_PATTERN.test(year)
  ) {
    return false;
  }
  return Number(first) <= 12 && Number(second) <= 12;
};

const normalizeHeader = (value: string) =>
  value.trim().toLowerCase().replace(/[_-]+/g, " ").replace(/\s+/g, " ");

const detectDelimiter = (text: string) => {
  const candidates = [",", "\t", ";"] as const;
  const firstRecord = text.split(/\r?\n/, 1)[0] ?? "";
  let quoteOpen = false;
  const counts = new Map(candidates.map((candidate) => [candidate, 0]));

  for (let index = 0; index < firstRecord.length; index += 1) {
    const character = firstRecord[index];
    if (character === '"') {
      if (quoteOpen && firstRecord[index + 1] === '"') {
        index += 1;
      } else {
        quoteOpen = !quoteOpen;
      }
      continue;
    }
    if (!quoteOpen && counts.has(character as (typeof candidates)[number])) {
      const candidate = character as (typeof candidates)[number];
      counts.set(candidate, (counts.get(candidate) ?? 0) + 1);
    }
  }

  return candidates.reduce((best, candidate) =>
    (counts.get(candidate) ?? 0) > (counts.get(best) ?? 0) ? candidate : best,
  );
};

const makeUniqueHeaders = (rawHeaders: string[]) => {
  const seen = new Map<string, number>();
  return rawHeaders.map((header, index) => {
    const base = header.trim() || `Column ${index + 1}`;
    const key = normalizeHeader(base);
    const count = (seen.get(key) ?? 0) + 1;
    seen.set(key, count);
    return count === 1 ? base : `${base} (${count})`;
  });
};

export const parseDelimitedText = (input: string) => {
  const text = input.replace(/^\uFEFF/, "");
  const delimiter = detectDelimiter(text);
  const records: string[][] = [];
  let record: string[] = [];
  let cell = "";
  let quoted = false;

  const pushCell = () => {
    record.push(cell.slice(0, MAX_CELL_CHARACTERS));
    cell = "";
  };
  const pushRecord = () => {
    pushCell();
    if (record.some((value) => value.trim())) records.push(record);
    record = [];
  };

  for (let index = 0; index < text.length; index += 1) {
    const character = text[index];
    if (character === '"') {
      if (quoted && text[index + 1] === '"') {
        cell += '"';
        index += 1;
      } else {
        quoted = !quoted;
      }
      continue;
    }
    if (!quoted && character === delimiter) {
      pushCell();
      continue;
    }
    if (!quoted && (character === "\n" || character === "\r")) {
      if (character === "\r" && text[index + 1] === "\n") index += 1;
      pushRecord();
      continue;
    }
    cell += character;
  }

  if (cell || record.length > 0) pushRecord();
  if (quoted) throw new Error("The file contains an unclosed quoted value.");
  if (records.length < 2) {
    throw new Error(
      "The file must contain a header and at least one task row.",
    );
  }

  const headers = makeUniqueHeaders(records[0].slice(0, MAX_COLUMNS));
  if (headers.length === 0)
    throw new Error("The file has no readable columns.");

  const rows = records
    .slice(1, IMPORT_MAX_TASKS + 1)
    .map((values) =>
      Object.fromEntries(
        headers.map((header, index) => [header, (values[index] ?? "").trim()]),
      ),
    );

  return {
    columns: headers,
    delimiter,
    rows,
    truncated: records.length - 1 > IMPORT_MAX_TASKS,
  };
};

export const inferImportMapping = (columns: string[]): ImportMapping => {
  const normalized = columns.map((column) => ({
    column,
    normalized: normalizeHeader(column),
  }));
  const used = new Set<string>();

  const find = (field: keyof ImportMapping) => {
    const aliases = FIELD_ALIASES[field];
    const exact = normalized.find(
      ({ column, normalized: value }) =>
        !used.has(column) && aliases.includes(value),
    );
    const partial = normalized.find(
      ({ column, normalized: value }) =>
        !used.has(column) &&
        aliases.some(
          (alias) =>
            value.startsWith(`${alias} `) || value.endsWith(` ${alias}`),
        ),
    );
    const match = exact ?? partial;
    if (match) used.add(match.column);
    return match?.column ?? null;
  };

  const mapping: ImportMapping = {
    title: find("title"),
    description: find("description"),
    status: find("status"),
    priority: find("priority"),
    assigneeEmail: find("assigneeEmail"),
    startDate: find("startDate"),
    endDate: find("endDate"),
    sourceId: find("sourceId"),
  };

  mapping.title ??= columns.at(0) ?? null;
  return mapping;
};

const normalizePriority = (value: string): ImportPriority => {
  const normalized = value.trim().toLowerCase();
  if (["urgent", "highest", "blocker", "critical", "p0"].includes(normalized)) {
    return "Urgent";
  }
  if (["high", "major", "p1"].includes(normalized)) return "High";
  if (["medium", "normal", "p2"].includes(normalized)) return "Medium";
  if (["low", "lowest", "minor", "trivial", "p3", "p4"].includes(normalized)) {
    return "Low";
  }
  return "No Priority";
};

const normalizeDate = (value: string) => {
  const trimmed = value.trim();
  if (!trimmed) return null;
  if (isAmbiguousNumericDate(trimmed)) return null;
  for (const pattern of DATE_PATTERNS) {
    const parsed = parse(trimmed, pattern, new Date(2000, 0, 1));
    if (isValid(parsed)) return format(parsed, "yyyy-MM-dd");
  }
  if (/^\d{4}-\d{2}-\d{2}$/.test(trimmed)) return null;
  const nativeDate = new Date(trimmed);
  return isValid(nativeDate) ? format(nativeDate, "yyyy-MM-dd") : null;
};

const getMappedValue = (row: Record<string, string>, column: string | null) =>
  column ? row[column] ?? "" : "";

export const mapRowsToImportTasks = (
  rows: Record<string, string>[],
  mapping: ImportMapping,
): ImportTask[] => {
  const sourceIdCounts = new Map<string, number>();

  return rows.flatMap((row, index) => {
    const title = getMappedValue(row, mapping.title).trim();
    if (!title) return [];
    const sourceIdBase =
      getMappedValue(row, mapping.sourceId).trim() || `row-${index + 2}`;
    const sourceIdCount = (sourceIdCounts.get(sourceIdBase) ?? 0) + 1;
    sourceIdCounts.set(sourceIdBase, sourceIdCount);
    const sourceId =
      sourceIdCount === 1 ? sourceIdBase : `${sourceIdBase}#${index + 2}`;
    const assignee = getMappedValue(row, mapping.assigneeEmail).trim();
    const status = getMappedValue(row, mapping.status).trim();

    return [
      {
        sourceId,
        title: title.slice(0, 255),
        description: getMappedValue(row, mapping.description).slice(0, 20_000),
        status: status || null,
        priority: normalizePriority(getMappedValue(row, mapping.priority)),
        assigneeEmail: assignee || null,
        startDate: normalizeDate(getMappedValue(row, mapping.startDate)),
        endDate: normalizeDate(getMappedValue(row, mapping.endDate)),
      },
    ];
  });
};

const isJiraExport = (columns: string[], mapping: ImportMapping) => {
  const normalized = new Set(columns.map(normalizeHeader));
  const sourceIdColumn = mapping.sourceId
    ? normalizeHeader(mapping.sourceId)
    : "";
  return Boolean(
    mapping.title &&
      (sourceIdColumn === "issue key" || sourceIdColumn === "key") &&
      (normalized.has("issue key") ||
        normalized.has("project key") ||
        normalized.has("issue type")),
  );
};

export const createDelimitedImportDraft = ({
  fileHash,
  fileName,
  text,
}: {
  fileHash: string;
  fileName: string;
  text: string;
}): ImportDraft => {
  const parsed = parseDelimitedText(text);
  const mapping = inferImportMapping(parsed.columns);
  const sourceType: ImportSourceType = isJiraExport(parsed.columns, mapping)
    ? "jira_csv"
    : "csv";
  const tasks = mapRowsToImportTasks(parsed.rows, mapping);
  const ambiguousDateCount = parsed.rows.reduce((count, row) => {
    const values = [
      getMappedValue(row, mapping.startDate),
      getMappedValue(row, mapping.endDate),
    ];
    return count + values.filter(isAmbiguousNumericDate).length;
  }, 0);
  const warnings = [
    ...(parsed.truncated
      ? [`Only the first ${IMPORT_MAX_TASKS} rows are included in this import.`]
      : []),
    ...(tasks.length < parsed.rows.length
      ? [
          `${parsed.rows.length - tasks.length} rows have no mapped title and will be skipped unless you change the mapping.`,
        ]
      : []),
    ...(ambiguousDateCount > 0
      ? [
          `${ambiguousDateCount} ambiguous numeric ${ambiguousDateCount === 1 ? "date was" : "dates were"} left blank. Use YYYY-MM-DD or an unambiguous date format before importing ${ambiguousDateCount === 1 ? "it" : "them"}.`,
        ]
      : []),
  ];

  const analysis: ImportAnalysis = {
    sourceType,
    summary:
      sourceType === "jira_csv"
        ? `Recognized a Jira export with ${tasks.length} importable issues.`
        : `Found ${tasks.length} importable tasks across ${parsed.columns.length} columns.`,
    warnings,
    mapping,
    tasks,
  };

  return {
    ...analysis,
    columns: parsed.columns,
    fileHash,
    fileName,
    rows: parsed.rows,
  };
};
