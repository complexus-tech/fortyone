import type {
  ImportSourceComment,
  ImportSourceCustomField,
  ImportTask,
} from "./schema";

type RecordValue = Record<string, unknown>;
const isRecord = (value: unknown): value is RecordValue =>
  typeof value === "object" && value !== null && !Array.isArray(value);
const records = (value: unknown) =>
  Array.isArray(value) ? value.filter(isRecord) : [];
const text = (value: unknown, limit = 255) =>
  typeof value === "string" ? value.trim().slice(0, limit) : "";

export const readTrelloCustomFields = (
  board: RecordValue,
  teamSourceId: string,
) => {
  const fields: ImportSourceCustomField[] = [];
  let unsupported = 0;
  for (const field of records(board.customFields)) {
    const sourceId = text(field.id, 300);
    const name = text(field.name, 100);
    const type = text(field.type);
    if (
      !sourceId ||
      !name ||
      !["text", "number", "date", "list", "checkbox"].includes(type)
    ) {
      unsupported += 1;
      continue;
    }
    fields.push({
      sourceId,
      name,
      currency: null,
      teamSourceId,
      type:
        type === "list" || type === "checkbox"
          ? "select"
          : (type as "text" | "number" | "date"),
      options:
        type === "checkbox"
          ? [
              { sourceId: "true", name: "Checked" },
              { sourceId: "false", name: "Unchecked" },
            ]
          : records(field.options).flatMap((option) => {
              const value = isRecord(option.value) ? option.value : {};
              const optionId = text(option.id, 300);
              const optionName = text(value.text, 100);
              return optionId && optionName
                ? [{ sourceId: optionId, name: optionName }]
                : [];
            }),
    });
  }
  return { fields, unsupported };
};

export const readTrelloCardValues = (
  card: RecordValue,
  fields: ImportSourceCustomField[],
) => {
  const fieldsById = new Map(fields.map((field) => [field.sourceId, field]));
  return records(card.customFieldItems).flatMap((item) => {
    const sourceFieldId = text(item.idCustomField, 300);
    const field = fieldsById.get(sourceFieldId);
    if (!field) return [];
    const sourceValue = isRecord(item.value) ? item.value : {};
    let value = text(sourceValue.text, 10_000);
    if (field.type === "number") value = text(sourceValue.number, 100);
    if (field.type === "date") value = text(sourceValue.date, 100).slice(0, 10);
    if (field.type === "select")
      value = text(item.idValue, 300) || text(sourceValue.checked, 10);
    return value ? [{ sourceFieldId, value }] : [];
  });
};

export const readTrelloComments = (board: RecordValue) => {
  const byCardId = new Map<string, ImportSourceComment[]>();
  let omitted = 0;
  for (const action of records(board.actions)) {
    if (action.type !== "commentCard") continue;
    const data = isRecord(action.data) ? action.data : {};
    const card = isRecord(data.card) ? data.card : {};
    const member = isRecord(action.memberCreator) ? action.memberCreator : {};
    const cardId = text(card.id, 300);
    const sourceId = text(action.id, 300);
    const content = typeof data.text === "string" ? data.text.trim() : "";
    const current = byCardId.get(cardId) ?? [];
    if (content.length > 20_000)
      throw new Error(
        "A Trello comment exceeds the 20,000 character source limit. Split this source comment before importing so its content is preserved.",
      );
    if (!cardId || !sourceId || !content || current.length >= 5_000) {
      omitted += 1;
      continue;
    }
    const date = text(action.date, 100);
    const parsedDate = new Date(date);
    current.push({
      sourceId,
      content,
      authorName:
        text(member.fullName) ||
        text(member.username) ||
        "Unknown Trello member",
      createdAt:
        date && !Number.isNaN(parsedDate.getTime())
          ? parsedDate.toISOString()
          : null,
    });
    byCardId.set(cardId, current);
  }
  for (const comments of byCardId.values())
    comments.sort(
      (a, b) =>
        (a.createdAt ?? "").localeCompare(b.createdAt ?? "") ||
        a.sourceId.localeCompare(b.sourceId),
    );
  return { byCardId, omitted };
};

export const createTrelloChecklistTasks = (
  parent: ImportTask,
  checklists: RecordValue[],
): ImportTask[] =>
  checklists.flatMap((checklist, index) =>
    records(checklist.checkItems).map((item, itemIndex) => {
      const checklistId = text(checklist.id, 200) || `list-${index + 1}`;
      const itemId = text(item.id, 200) || `item-${itemIndex + 1}`;
      const complete = item.state === "complete";
      return {
        ...parent,
        sourceId:
          `trello-check:${parent.sourceId}:${checklistId}:${itemId}`.slice(
            0,
            300,
          ),
        title: text(item.name) || "Untitled checklist item",
        description: `Checklist: ${text(checklist.name) || "Checklist"}`,
        parentSourceId: parent.sourceId,
        status: complete ? "Completed" : "To do",
        statusCategory: complete ? "completed" : "unstarted",
        assigneeName: null,
        assigneeEmail: null,
        assigneePersonSourceId: text(item.idMember, 300) || null,
        collaboratorPersonSourceIds: [],
        labelSourceIds: [],
        links: [],
        comments: [],
        customFieldValues: [],
        startDate: null,
        endDate: null,
      };
    }),
  );
