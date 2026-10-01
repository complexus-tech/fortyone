import { createDelimitedImportDraft } from "./csv";
import { createJsonImportDraft } from "./json";
import { splitSourceComment } from "./import-comments";
import { getCanonicalImportEffort } from "./canonical-effort";
import { workExportSchema } from "./backup-format";
import { createEmptyImportEntityCollections, importTaskSchema } from "./schema";

const task = importTaskSchema.parse({
  sourceId: "task-1",
  title: "Presentation title",
  description: "Presentation description",
  status: null,
  statusCategory: "completed",
  priority: "No Priority",
  estimateValue: null,
  estimatedDurationMinutes: null,
  minimumFocusBlockMinutes: null,
  assigneeEmail: null,
  assigneeName: null,
  assigneePersonSourceId: null,
  collaboratorPersonSourceIds: [],
  teamSourceId: null,
  parentSourceId: null,
  objectiveSourceId: null,
  keyResultSourceId: null,
  sprintSourceId: null,
  labelSourceIds: [],
  associations: [],
  links: [],
  startDate: null,
  endDate: null,
});

describe("migration preservation", () => {
  it("keeps Trello source comment attribution and exact typed field strings", () => {
    const draft = createJsonImportDraft({
      fileHash: "a".repeat(64),
      fileName: "board.json",
      text: JSON.stringify({
        id: "board-1",
        prefs: { permissionLevel: "private" },
        lists: [{ id: "list-1", name: "To do" }],
        cards: [
          {
            id: "card-1",
            idList: "list-1",
            name: "Customer contract",
            customFieldItems: [
              {
                idCustomField: "value",
                value: { number: "9007199254740993.29" },
              },
              { idCustomField: "stage", idValue: "option-1" },
            ],
          },
        ],
        customFields: [
          { id: "value", name: "Contract value", type: "number" },
          {
            id: "stage",
            name: "Stage",
            type: "list",
            options: [{ id: "option-1", value: { text: "Won" } }],
          },
        ],
        actions: [
          {
            id: "comment-1",
            type: "commentCard",
            date: "2026-09-30T10:20:00Z",
            memberCreator: { fullName: "Source author" },
            data: { card: { id: "card-1" }, text: "Source decision" },
          },
        ],
      }),
    });
    expect(draft.customFields).toMatchObject([
      { name: "Contract value", type: "number" },
      {
        name: "Stage",
        type: "select",
        options: [{ sourceId: "option-1", name: "Won" }],
      },
    ]);
    expect(draft.tasks[0]?.customFieldValues).toEqual([
      { sourceFieldId: "value", value: "9007199254740993.29" },
      { sourceFieldId: "stage", value: "option-1" },
    ]);
    expect(draft.tasks[0]?.comments).toEqual([
      {
        sourceId: "comment-1",
        content: "Source decision",
        authorName: "Source author",
        createdAt: "2026-09-30T10:20:00.000Z",
      },
    ]);
  });
  it("accepts more than500 tasks but keeps the graph bounded", () => {
    const text = JSON.stringify(
      Array.from({ length: 1200 }, (_, index) => ({
        id: `task-${index}`,
        name: `Task ${index}`,
      })),
    );
    const draft = createJsonImportDraft({
      text,
      fileName: "large.json",
      fileHash: "a".repeat(64),
    });
    expect(draft.tasks).toHaveLength(1200);
  });
  it("restores canonical backup content without truncating native limits", () => {
    const description = "Work detail ".repeat(7000);
    const title = "T".repeat(450);
    const backup = {
      format: "fortyone-work-export",
      version: 1,
      generatedAt: "2026-10-01T12:00:00Z",
      scope: { teamSourceId: null },
      analysis: {
        sourceType: "json",
        sourceNamespace: "fortyone:workspace:origin",
        summary: "Backup",
        warnings: [],
        mapping: null,
        ...createEmptyImportEntityCollections(),
        tasks: [task],
      },
      customFields: [
        {
          sourceId: "value",
          name: "Historical value",
          type: "money",
          icon: "analytics",
          currency: "USD",
          teamSourceId: null,
          options: [],
          archivedAt: "2026-09-01T12:00:00Z",
        },
        {
          sourceId: "stage",
          name: "Stage",
          type: "select",
          icon: null,
          currency: null,
          teamSourceId: null,
          options: [
            {
              sourceId: "won",
              name: "Won",
              archivedAt: "2026-09-01T12:00:00Z",
            },
          ],
          archivedAt: null,
        },
      ],
      taskData: [
        {
          sourceId: "task-1",
          title,
          description,
          descriptionHTML: `<p>${description}</p>`,
          estimateValue: 13,
          estimatedDurationMinutes: 6000,
          minimumFocusBlockMinutes: 60,
          createdAt: "2025-01-01T12:00:00Z",
          comments: [
            {
              sourceId: "comment-1",
              content: "<p>Decision</p>",
              authorName: "Ada",
              createdAt: "2025-01-01T12:00:00Z",
              parentSourceId: null,
              format: "html",
            },
            {
              sourceId: "reply-1",
              content: "<p>Source reply</p>",
              authorName: "Grace",
              createdAt: "2025-01-02T12:00:00Z",
              parentSourceId: "comment-1",
              format: "html",
            },
          ],
          customFieldValues: [
            { sourceFieldId: "value", value: "9007199254740993.29" },
            { sourceFieldId: "stage", value: "won" },
          ],
        },
      ],
    };
    expect(workExportSchema.safeParse(backup).success).toBe(true);
    const draft = createJsonImportDraft({
      text: JSON.stringify(backup),
      fileHash: "a".repeat(64),
      fileName: "backup.json",
    });
    expect(draft.tasks[0]?.canonical).toMatchObject({
      title,
      description,
      estimateValue: 13,
      estimatedDurationMinutes: 6000,
    });
    expect(getCanonicalImportEffort(draft.tasks[0])).toEqual({
      estimateValue: undefined,
      estimatedDurationMinutes: undefined,
      minimumFocusBlockMinutes: undefined,
      unsupported: true,
    });
    expect(
      draft.warnings.some((warning) =>
        warning.includes("source effort outside"),
      ),
    ).toBe(true);
    expect(draft.customFields?.[0]?.archivedAt).toBe("2026-09-01T12:00:00Z");
    expect(draft.customFields?.[0]?.icon).toBe("analytics");
    expect(draft.customFields?.[1]?.icon).toBeNull();
    expect(draft.customFields?.[1]?.options[0]?.archivedAt).toBe(
      "2026-09-01T12:00:00Z",
    );
    expect(draft.tasks[0]?.customFieldValues).toEqual([
      { sourceFieldId: "value", value: "9007199254740993.29" },
      { sourceFieldId: "stage", value: "won" },
    ]);
    expect(draft.tasks[0]?.comments?.[1]).toMatchObject({
      parentSourceId: "comment-1",
      content: "<p>Source reply</p>",
    });
    expect(draft.tasks[0]?.comments?.[0]).toMatchObject({
      format: "html",
      authorName: "Ada",
      parentSourceId: null,
    });
  });
  it("splits large plain comments into stable safe continuation IDs without losing text", () => {
    const original = {
      sourceId: "comment-1",
      content: "<&>😀".repeat(1100),
      authorName: "Source member",
      createdAt: null,
    };
    const parts = splitSourceComment(original);
    expect(parts.map((part) => part.content).join("")).toBe(original.content);
    expect(parts[0]?.sourceId).toBe("comment-1");
    expect(parts[1]).toMatchObject({
      sourceId: "comment-1:part:2",
      parentSourceId: "comment-1",
    });
    expect(parts.every((part) => Array.from(part.content).length <= 1000)).toBe(
      true,
    );
  });
  it("preserves oversized HTML as full attributed source text instead of splitting tags", () => {
    const content = `<p>${"A & <strong>source note</strong>".repeat(400)}</p>`;
    const parts = splitSourceComment({
      sourceId: "html-1",
      content,
      format: "html",
      authorName: "Ada",
      createdAt: null,
    });
    expect(parts.length).toBeGreaterThan(1);
    expect(parts.map((part) => part.content).join("")).toBe(content);
    expect(
      parts.every(
        (part) =>
          part.format === "text" && Array.from(part.content).length <= 1000,
      ),
    ).toBe(true);
    expect(parts[1].parentSourceId).toBe("html-1");
  });
  it("maps documented Shortcut CSV states and checklist syntax explicitly", () => {
    const draft = createDelimitedImportDraft({
      text: 'id,name,workflow_id,is_completed,owners,labels,tasks\n42,Ship work,2,TRUE,ada@example.com,Customer,"[X] Verify;[ ] Publish"',
      fileHash: "a".repeat(64),
      fileName: "shortcut.csv",
    });
    expect(draft.sourceMetadata?.platform).toBe("shortcut");
    expect(draft.tasks).toHaveLength(3);
    expect(draft.tasks[0]).toMatchObject({
      sourceId: "42",
      statusCategory: "completed",
    });
    expect(draft.tasks[1]).toMatchObject({
      parentSourceId: "42",
      statusCategory: "completed",
    });
    expect(draft.people[0]?.email).toBe("ada@example.com");
    expect(
      draft.warnings.some((warning) => warning.includes("refreshed export")),
    ).toBe(true);
  });
});
