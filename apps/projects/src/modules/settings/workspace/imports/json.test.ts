/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { createJsonImportDraft } from "./json";

describe("work import JSON parsing", () => {
  it("creates a vendor-neutral preview from a nested board export", () => {
    const draft = createJsonImportDraft({
      fileHash: "trello-hash",
      fileName: "product-board.json",
      text: JSON.stringify({
        id: " board/65 ",
        actions: [{ type: "commentCard" }],
        cards: [
          {
            attachments: [{ id: "attachment-1" }],
            closed: true,
            desc: "Keep the migration safe.",
            due: "2026-09-15T12:00:00.000Z",
            id: "65a1234567890abcdef12345",
            idList: "list-doing",
            idMembers: ["member-1"],
            labels: [{ name: "P1 High" }],
            name: "Migrate the product board",
            start: "2026-09-10T09:00:00.000Z",
          },
        ],
        checklists: [
          {
            checkItems: [
              { name: "Export cards", state: "complete" },
              { name: "Verify owners", state: "incomplete" },
            ],
            idCard: "65a1234567890abcdef12345",
            name: "Migration",
          },
        ],
        lists: [{ id: "list-doing", name: "In progress" }],
        members: [{ email: "owner@example.com", id: "member-1" }],
        name: "Product",
        prefs: { permissionLevel: "private" },
      }),
    });

    expect(draft.sourceType).toBe("json");
    expect(draft.sourceNamespace).toBe("trello:board:board%2F65");
    expect(draft.summary).toContain("Semantic mapping");
    expect(draft.tasks).toEqual([
      expect.objectContaining({
        assigneeEmail: null,
        endDate: "2026-09-15",
        priority: "No Priority",
        sourceId: "65a1234567890abcdef12345",
        startDate: "2026-09-10",
        status: null,
        title: "Migrate the product board",
      }),
    ]);
    expect(draft.tasks[0]?.description).toContain("Keep the migration safe.");
    expect(draft.rows[0]?.idMembers).toBe('["member-1"]');
    expect(draft.rows[0]?.labels).toBe('[{"name":"P1 High"}]');
    expect(draft).toMatchObject({
      teams: [],
      people: [],
      labels: [],
      strategicPillars: [],
      objectives: [],
      keyResults: [],
      sprints: [],
    });
    expect(draft.tasks[0]).toMatchObject({
      statusCategory: null,
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
    });
  });

  it("maps a generic JSON task collection and keeps nested values reviewable", () => {
    const draft = createJsonImportDraft({
      fileHash: "json-hash",
      fileName: "tasks.json",
      text: JSON.stringify({
        tasks: [
          {
            id: "task-1",
            metadata: { source: "legacy" },
            owner: "person@example.com",
            priority: "urgent",
            title: "Move the backlog",
          },
        ],
      }),
    });

    expect(draft.sourceType).toBe("json");
    expect(draft.sourceNamespace).toBeNull();
    expect(draft.tasks[0]).toEqual(
      expect.objectContaining({
        priority: "Urgent",
        sourceId: "task-1",
        title: "Move the backlog",
      }),
    );
    expect(draft.rows[0]?.metadata).toBe('{"source":"legacy"}');
  });

  it("keeps a non-task JSON graph available for semantic analysis", () => {
    const draft = createJsonImportDraft({
      fileHash: "portfolio-hash",
      fileName: "portfolio.json",
      text: JSON.stringify({
        id: 42,
        projects: [{ id: "project-1", name: "Improve activation" }],
        teams: [{ id: "team-1", name: "Growth" }],
      }),
    });

    expect(draft).toMatchObject({
      sourceType: "json",
      sourceNamespace: null,
      teams: [],
      strategicPillars: [],
      objectives: [],
      tasks: [],
      columns: [],
      rows: [],
    });
    expect(draft.summary).toContain("teams, people, strategic pillars");
    expect(draft.warnings).toContain(
      "No standard task collection was found for the initial preview. AI analysis can still map supported objects from the complete JSON document.",
    );
  });

  it("keeps derived namespaces control-free and within the UTF-8 byte limit", () => {
    const controlDraft = createJsonImportDraft({
      fileHash: "control-hash",
      fileName: "control.json",
      text: JSON.stringify({
        cards: [{ idList: "list-1" }],
        id: "board\u0000id",
        lists: [{ id: "list-1" }],
        prefs: {},
      }),
    });
    const longUnicodeDraft = createJsonImportDraft({
      fileHash: "unicode-hash",
      fileName: "unicode.json",
      text: JSON.stringify({
        cards: [{ idList: "list-1" }],
        id: "界".repeat(100),
        lists: [{ id: "list-1" }],
        prefs: {},
      }),
    });

    expect(controlDraft.sourceNamespace).toBe("trello:board:board%00id");
    expect(
      [...(controlDraft.sourceNamespace ?? "")].some((character) => {
        const code = character.charCodeAt(0);
        return code <= 31 || code === 127;
      }),
    ).toBe(false);
    expect(
      Buffer.byteLength(controlDraft.sourceNamespace ?? "", "utf8"),
    ).toBeLessThanOrEqual(300);
    expect(longUnicodeDraft.sourceNamespace).toMatch(
      /^trello:board:.+~[0-9a-f]{16}$/u,
    );
    expect(
      Buffer.byteLength(longUnicodeDraft.sourceNamespace ?? "", "utf8"),
    ).toBeLessThanOrEqual(300);
  });

  it("does not assign Trello identity to a generic cards-and-lists graph", () => {
    const draft = createJsonImportDraft({
      fileHash: "kanban-hash",
      fileName: "kanban.json",
      text: JSON.stringify({
        cards: [{ id: "card-1", idList: "list-1", name: "Generic card" }],
        id: "generic-board",
        lists: [{ id: "list-1", name: "Doing" }],
      }),
    });

    expect(draft.sourceNamespace).toBeNull();
  });

  it("rejects malformed JSON with an actionable error", () => {
    expect(() =>
      createJsonImportDraft({
        fileHash: "json-hash",
        fileName: "tasks.json",
        text: "{not-json}",
      }),
    ).toThrow("The JSON file is not valid JSON.");
  });

  it("rejects primitive JSON values", () => {
    expect(() =>
      createJsonImportDraft({
        fileHash: "json-hash",
        fileName: "value.json",
        text: '"not an export"',
      }),
    ).toThrow("The JSON file must contain an object or array.");
  });
});
