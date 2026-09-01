import { z } from "zod";
import type {
  ImportAnalysis,
  ImportDraft,
  ImportLabel,
  ImportPerson,
  ImportStatusCategory,
  ImportTask,
} from "./schema";
import {
  createEmptyImportEntityCollections,
  IMPORT_MAX_LABELS,
  IMPORT_MAX_PEOPLE,
  IMPORT_MAX_TASKS,
  normalizeImportTaskLinks,
} from "./schema";
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
const TRELLO_MEMBER_EMAIL_SCHEMA = z.email().max(320);

type JsonRecord = Record<string, unknown>;

type RecordTable = {
  columns: string[];
  rows: Record<string, string>[];
  sourceRecords: JsonRecord[];
};

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

const createRecordTable = (records: JsonRecord[]): RecordTable => {
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
  return { columns, rows, sourceRecords };
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

const normalizeSourceId = (value: unknown, fallback: string) => {
  let sourceId = "";
  if (typeof value === "string") sourceId = value.normalize("NFKC").trim();
  else if (typeof value === "number" && Number.isSafeInteger(value)) {
    sourceId = String(value);
  }
  sourceId = [...sourceId]
    .filter((character) => {
      const code = character.charCodeAt(0);
      return code > 31 && code !== 127;
    })
    .join("");
  if (!sourceId) return fallback;
  if (sourceId.length <= 300) return sourceId;

  const suffix = `~${stableTextHash(sourceId)}`;
  return `${sourceId.slice(0, 300 - suffix.length)}${suffix}`;
};

const normalizeText = (value: unknown, maxLength: number) =>
  typeof value === "string"
    ? value.normalize("NFKC").trim().slice(0, maxLength)
    : "";

const normalizeInlineText = (value: unknown, fallback: string) =>
  normalizeText(value, MAX_CELL_CHARACTERS).replace(/\s+/g, " ") || fallback;

const normalizeTrelloMemberEmail = (value: unknown) => {
  if (typeof value !== "string") return null;
  const normalized = value.normalize("NFKC").trim().toLowerCase();
  const result = TRELLO_MEMBER_EMAIL_SCHEMA.safeParse(normalized);
  return result.success ? result.data : null;
};

const normalizeTrelloDate = (value: unknown) => {
  if (typeof value !== "string") return null;
  const trimmed = value.trim();
  const datePart = trimmed.slice(0, 10);
  if (!/^\d{4}-\d{2}-\d{2}$/.test(datePart)) return null;
  if (trimmed.length > 10 && trimmed.charAt(10) !== "T") return null;
  const date = new Date(`${datePart}T00:00:00.000Z`);
  if (Number.isNaN(date.getTime())) return null;
  return date.toISOString().slice(0, 10) === datePart ? datePart : null;
};

const getRecordArray = (value: unknown) =>
  Array.isArray(value) ? value.filter(isRecord) : [];

const formatEntityCount = (count: number, singular: string) =>
  `${count} ${count === 1 ? singular : `${singular}s`}`;

const getSourceIdArray = (value: unknown) => {
  if (!Array.isArray(value)) return [];
  const sourceIds = new Set<string>();
  for (const candidate of value) {
    if (typeof candidate !== "string" && typeof candidate !== "number") {
      continue;
    }
    const sourceId = normalizeSourceId(candidate, "");
    if (sourceId) sourceIds.add(sourceId);
  }
  return [...sourceIds];
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

const getTrelloStatusCategory = (
  status: string,
  dueComplete: boolean,
): ImportStatusCategory => {
  const normalized = status.toLocaleLowerCase();
  if (
    dueComplete ||
    /(?:^|\s)(?:done|complete|completed|shipped|resolved)(?:$|\s)/.test(
      normalized,
    )
  ) {
    return "completed";
  }
  if (
    /wont do|won't do|cancelled|canceled|cannot reproduce|rejected/.test(
      normalized,
    )
  ) {
    return "cancelled";
  }
  if (/on hold|paused|blocked|waiting/.test(normalized)) return "paused";
  if (/doing|in progress|testing|review|active/.test(normalized)) {
    return "started";
  }
  if (/backlog|icebox|inbox/.test(normalized)) return "backlog";
  return "unstarted";
};

const formatTrelloChecklist = (checklist: JsonRecord) => {
  const items = getRecordArray(checklist.checkItems);
  if (items.length === 0) return { itemCount: 0, markdown: "" };
  const title = normalizeInlineText(checklist.name, "Checklist");
  const itemLines = items.map((item) => {
    const checked = item.state === "complete" ? "x" : " ";
    const name = normalizeInlineText(item.name, "Untitled checklist item");
    return `- [${checked}] ${name}`;
  });
  return {
    itemCount: items.length,
    markdown: [`### ${title}`, ...itemLines].join("\n"),
  };
};

const createTrelloImportDraft = ({
  board,
  fileHash,
  fileName,
  sourceNamespace,
}: {
  board: JsonRecord;
  fileHash: string;
  fileName: string;
  sourceNamespace: string;
}): ImportDraft => {
  const cardRecords = getRecordArray(board.cards);
  const table = createRecordTable(cardRecords);
  const boardSourceId = normalizeSourceId(
    board.id,
    `trello-board-${stableTextHash(sourceNamespace)}`,
  );
  const boardName =
    normalizeText(board.name, 255) ||
    normalizeText(fileName.replace(/\.json$/i, ""), 255) ||
    "Trello import";
  const boardDescription =
    normalizeText(board.desc, MAX_CELL_CHARACTERS) || null;
  const preferences = isRecord(board.prefs) ? board.prefs : {};
  const permissionLevel = normalizeText(preferences.permissionLevel, 50);
  const teams: ImportAnalysis["teams"] = [
    {
      sourceId: boardSourceId,
      name: boardName,
      code: null,
      color: null,
      description: boardDescription,
      isPrivate: permissionLevel === "private",
    },
  ];

  const listBySourceId = new Map<string, JsonRecord>();
  for (const [index, list] of getRecordArray(board.lists).entries()) {
    const sourceId = normalizeSourceId(list.id, `trello-list-${index + 1}`);
    listBySourceId.set(sourceId, list);
  }

  const checklistsByCardSourceId = new Map<string, JsonRecord[]>();
  for (const checklist of getRecordArray(board.checklists)) {
    const cardSourceId = normalizeSourceId(checklist.idCard, "");
    if (!cardSourceId) continue;
    const current = checklistsByCardSourceId.get(cardSourceId) ?? [];
    current.push(checklist);
    checklistsByCardSourceId.set(cardSourceId, current);
  }

  const memberRecordsBySourceId = new Map<string, JsonRecord>();
  for (const [index, member] of getRecordArray(board.members).entries()) {
    const sourceId = normalizeSourceId(member.id, `trello-member-${index + 1}`);
    memberRecordsBySourceId.set(sourceId, member);
  }
  const membershipByPersonSourceId = new Map<string, boolean>();
  for (const membership of getRecordArray(board.memberships)) {
    const sourceId = normalizeSourceId(membership.idMember, "");
    if (!sourceId) continue;
    membershipByPersonSourceId.set(
      sourceId,
      membership.deactivated !== true && membership.unconfirmed !== true,
    );
    if (!memberRecordsBySourceId.has(sourceId)) {
      memberRecordsBySourceId.set(sourceId, {});
    }
  }
  const assignedPersonSourceIds = new Set<string>();
  for (const card of table.sourceRecords) {
    for (const sourceId of getSourceIdArray(card.idMembers)) {
      assignedPersonSourceIds.add(sourceId);
      if (!memberRecordsBySourceId.has(sourceId)) {
        memberRecordsBySourceId.set(sourceId, {});
      }
    }
  }
  const hasExplicitMemberships = membershipByPersonSourceId.size > 0;
  const people: ImportPerson[] = [...memberRecordsBySourceId]
    .slice(0, IMPORT_MAX_PEOPLE)
    .map(([sourceId, member]) => {
      const name =
        normalizeText(member.fullName, 255) ||
        normalizeText(member.username, 255) ||
        null;
      const isTeamMember = membershipByPersonSourceId.has(sourceId)
        ? membershipByPersonSourceId.get(sourceId) === true
        : !hasExplicitMemberships || assignedPersonSourceIds.has(sourceId);
      return {
        sourceId,
        name,
        email: normalizeTrelloMemberEmail(member.email),
        teamSourceIds: isTeamMember ? [boardSourceId] : [],
      };
    });
  const retainedPeopleSourceIds = new Set(
    people.map((person) => person.sourceId),
  );
  const peopleBySourceId = new Map(
    people.map((person) => [person.sourceId, person]),
  );

  const labelBySourceId = new Map<string, ImportLabel>();
  const registerLabel = (label: JsonRecord) => {
    const rawName = normalizeText(label.name, 100);
    const color = normalizeText(label.color, 100) || null;
    const sourceId = normalizeSourceId(
      label.id,
      `trello-label-${stableTextHash(`${rawName}\u0000${color ?? ""}`)}`,
    );
    if (!labelBySourceId.has(sourceId)) {
      const fallbackLabel = color
        ? `${color.charAt(0).toLocaleUpperCase()}${color.slice(1)} label`
        : `Trello label ${sourceId.slice(-6)}`;
      labelBySourceId.set(sourceId, {
        sourceId,
        name: (rawName || fallbackLabel).slice(0, 100),
        color,
        teamSourceId: boardSourceId,
      });
    }
    return sourceId;
  };
  for (const label of getRecordArray(board.labels)) registerLabel(label);
  for (const card of table.sourceRecords) {
    for (const label of getRecordArray(card.labels)) registerLabel(label);
    for (const sourceId of getSourceIdArray(card.idLabels)) {
      registerLabel({ id: sourceId });
    }
  }
  const labels = [...labelBySourceId.values()].slice(0, IMPORT_MAX_LABELS);
  const retainedLabelSourceIds = new Set(labels.map((label) => label.sourceId));

  const sourceIdCounts = new Map<string, number>();
  const archivedTaskSourceIds: string[] = [];
  let nestedChecklistItemCount = 0;
  let truncatedDescriptionCount = 0;
  const tasks: ImportTask[] = table.sourceRecords.map((card, index) => {
    const sourceIdBase = normalizeSourceId(card.id, `trello-card-${index + 1}`);
    const duplicateCount = (sourceIdCounts.get(sourceIdBase) ?? 0) + 1;
    sourceIdCounts.set(sourceIdBase, duplicateCount);
    const sourceId =
      duplicateCount === 1
        ? sourceIdBase
        : normalizeSourceId(
            `${sourceIdBase}#${duplicateCount}`,
            `trello-card-${index + 1}`,
          );
    const checklistSections = (
      checklistsByCardSourceId.get(sourceIdBase) ?? []
    ).flatMap((checklist) => {
      const formatted = formatTrelloChecklist(checklist);
      nestedChecklistItemCount += formatted.itemCount;
      return formatted.markdown ? [formatted.markdown] : [];
    });
    const sourceDescription =
      typeof card.desc === "string" ? card.desc.trim() : "";
    const fullDescription = [sourceDescription, ...checklistSections]
      .filter(Boolean)
      .join("\n\n");
    if (fullDescription.length > MAX_CELL_CHARACTERS) {
      truncatedDescriptionCount += 1;
    }

    const listSourceId = normalizeSourceId(card.idList, "");
    const list = listBySourceId.get(listSourceId);
    if (card.closed === true || list?.closed === true) {
      archivedTaskSourceIds.push(sourceId);
    }
    const status = normalizeText(list?.name, 200) || null;
    const cardPersonSourceIds = getSourceIdArray(card.idMembers).filter(
      (memberSourceId) => retainedPeopleSourceIds.has(memberSourceId),
    );
    const assigneePersonSourceId = cardPersonSourceIds.at(0) ?? null;
    const assigneeName = assigneePersonSourceId
      ? peopleBySourceId.get(assigneePersonSourceId)?.name ?? null
      : null;
    const cardLabelSourceIds = new Set<string>();
    for (const label of getRecordArray(card.labels)) {
      const labelSourceId = registerLabel(label);
      if (retainedLabelSourceIds.has(labelSourceId)) {
        cardLabelSourceIds.add(labelSourceId);
      }
    }
    for (const labelSourceId of getSourceIdArray(card.idLabels)) {
      if (retainedLabelSourceIds.has(labelSourceId)) {
        cardLabelSourceIds.add(labelSourceId);
      }
    }
    const canonicalLinks = normalizeImportTaskLinks(
      [card.url, card.shortUrl].flatMap((value) => {
        const url = normalizeText(value, 500);
        return url ? [{ title: "Trello card", url }] : [];
      }),
    ).slice(0, 1);
    const attachmentLinks = getRecordArray(card.attachments).flatMap(
      (attachment) => {
        const url = normalizeText(attachment.url, 500);
        if (!url) return [];
        return [
          {
            title: normalizeText(attachment.name, 255) || "Trello attachment",
            url,
          },
        ];
      },
    );
    const links = normalizeImportTaskLinks([
      ...canonicalLinks,
      ...attachmentLinks,
    ]);

    return {
      sourceId,
      title:
        normalizeText(card.name, 255) || `Untitled Trello card ${index + 1}`,
      description: fullDescription.slice(0, MAX_CELL_CHARACTERS),
      status,
      statusCategory: getTrelloStatusCategory(
        status ?? "",
        card.dueComplete === true,
      ),
      priority: "No Priority",
      estimateValue: null,
      estimatedDurationMinutes: null,
      minimumFocusBlockMinutes: null,
      assigneeEmail: null,
      assigneeName,
      assigneePersonSourceId,
      collaboratorPersonSourceIds: cardPersonSourceIds.slice(1, 101),
      teamSourceId: boardSourceId,
      parentSourceId: null,
      objectiveSourceId: null,
      keyResultSourceId: null,
      sprintSourceId: null,
      labelSourceIds: [...cardLabelSourceIds].slice(0, 100),
      associations: [],
      links,
      startDate: normalizeTrelloDate(card.start),
      endDate: normalizeTrelloDate(card.due),
    };
  });

  const commentCount = getRecordArray(board.actions).filter(
    (action) => action.type === "commentCard",
  ).length;
  const warnings = [
    ...(cardRecords.length > IMPORT_MAX_TASKS
      ? [
          `Only the first ${IMPORT_MAX_TASKS} Trello cards are included in this import.`,
        ]
      : []),
    ...(memberRecordsBySourceId.size > IMPORT_MAX_PEOPLE
      ? [
          `Only the first ${IMPORT_MAX_PEOPLE} Trello members are included in this import.`,
        ]
      : []),
    ...(labelBySourceId.size > IMPORT_MAX_LABELS
      ? [
          `Only the first ${IMPORT_MAX_LABELS} Trello labels are included in this import.`,
        ]
      : []),
    ...(truncatedDescriptionCount > 0
      ? [
          `${truncatedDescriptionCount} Trello card descriptions were shortened to fit the import limit.`,
        ]
      : []),
    ...(commentCount > 0
      ? [
          `${formatEntityCount(commentCount, "Trello card comment")} cannot be imported because comment activity is not supported yet.`,
        ]
      : []),
  ];
  const analysis: ImportAnalysis = {
    sourceType: "json",
    sourceNamespace,
    summary: `Found ${formatEntityCount(tasks.length, "Trello card")}, ${formatEntityCount(people.length, "member")}, ${formatEntityCount(labels.length, "label")}, and ${formatEntityCount(nestedChecklistItemCount, "checklist item")}. Checklist items stay with their parent cards.`,
    warnings,
    mapping: null,
    teams,
    people,
    labels,
    strategicPillars: [],
    objectives: [],
    keyResults: [],
    sprints: [],
    tasks,
  };

  return {
    ...analysis,
    columns: table.columns,
    fileHash,
    fileName,
    rows: table.rows,
    sourceMetadata: {
      archivedTaskSourceIds,
      nestedChecklistItemCount,
      platform: "trello",
    },
  };
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
  const sourceNamespace = findSourceNamespace(value);
  if (sourceNamespace && isRecord(value)) {
    return createTrelloImportDraft({
      board: value,
      fileHash,
      fileName,
      sourceNamespace,
    });
  }
  const records = findTaskRecords(value);
  const { columns, rows } = createRecordTable(records);
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
    sourceNamespace: null,
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
