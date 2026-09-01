/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { createJsonImportDraft } from "./json";

describe("work import JSON parsing", () => {
  it("normalizes a Trello board into one task per card", () => {
    const draft = createJsonImportDraft({
      fileHash: "trello-hash",
      fileName: "product-board.json",
      text: JSON.stringify({
        id: " board/65 ",
        actions: [{ type: "commentCard" }],
        cards: [
          {
            attachments: [
              {
                id: "attachment-1",
                name: "Migration plan",
                url: "https://example.com/migration-plan.pdf",
              },
              {
                id: "attachment-unsafe",
                name: "Unsafe attachment",
                url: "ftp://example.com/unsafe-attachment",
              },
              {
                id: "attachment-duplicate",
                name: "Duplicate card link",
                url: "https://trello.com/c/card-1/migrate-the-product-board",
              },
            ],
            closed: true,
            desc: "Keep the migration safe.",
            due: "2026-09-15T12:00:00.000Z",
            dueComplete: false,
            id: "65a1234567890abcdef12345",
            idLabels: ["label-high"],
            idList: "list-doing",
            idMembers: ["member-1", "member-2"],
            labels: [{ color: "red", id: "label-high", name: "P1 High" }],
            name: "Migrate the product board",
            shortUrl: "https://trello.com/c/card-1",
            start: "2026-09-10T09:00:00.000Z",
            url: "https://trello.com/c/card-1/migrate-the-product-board",
          },
        ],
        checklists: [
          {
            checkItems: [
              {
                id: "check-item-1",
                name: "Export cards",
                state: "complete",
              },
              {
                id: "check-item-2",
                name: "Verify owners",
                state: "incomplete",
              },
            ],
            id: "checklist-1",
            idCard: "65a1234567890abcdef12345",
            name: "Migration",
          },
        ],
        lists: [{ id: "list-doing", name: "In progress" }],
        members: [
          {
            email: " Owner@Example.COM ",
            fullName: "Owner One",
            id: "member-1",
            username: "owner-one",
          },
          {
            email: "owner-two",
            fullName: "Owner Two",
            id: "member-2",
            username: "owner-two",
          },
        ],
        memberships: [
          {
            deactivated: false,
            idMember: "member-1",
            unconfirmed: false,
          },
          {
            deactivated: false,
            idMember: "member-2",
            unconfirmed: false,
          },
        ],
        name: "Product",
        prefs: { permissionLevel: "private" },
      }),
    });

    expect(draft.sourceType).toBe("json");
    expect(draft.sourceNamespace).toBe("trello:board:board%2F65");
    expect(draft.summary).toBe(
      "Found 1 Trello card, 2 members, 1 label, and 2 checklist items. Checklist items stay with their parent cards.",
    );
    expect(draft.mapping).toBeNull();
    expect(draft.tasks).toEqual([
      {
        assigneeEmail: null,
        assigneeName: "Owner One",
        assigneePersonSourceId: "member-1",
        associations: [],
        collaboratorPersonSourceIds: ["member-2"],
        description:
          "Keep the migration safe.\n\n### Migration\n- [x] Export cards\n- [ ] Verify owners",
        endDate: "2026-09-15",
        estimateValue: null,
        estimatedDurationMinutes: null,
        keyResultSourceId: null,
        labelSourceIds: ["label-high"],
        links: [
          {
            title: "Trello card",
            url: "https://trello.com/c/card-1/migrate-the-product-board",
          },
          {
            title: "Migration plan",
            url: "https://example.com/migration-plan.pdf",
          },
        ],
        minimumFocusBlockMinutes: null,
        objectiveSourceId: null,
        parentSourceId: null,
        priority: "No Priority",
        sourceId: "65a1234567890abcdef12345",
        sprintSourceId: null,
        startDate: "2026-09-10",
        status: "In progress",
        statusCategory: "started",
        teamSourceId: "board/65",
        title: "Migrate the product board",
      },
    ]);
    expect(draft.tasks).toHaveLength(1);
    expect(draft.rows[0]?.idMembers).toBe('["member-1","member-2"]');
    expect(draft.rows[0]?.labels).toBe(
      '[{"color":"red","id":"label-high","name":"P1 High"}]',
    );
    expect(draft).toMatchObject({
      teams: [
        {
          sourceId: "board/65",
          name: "Product",
          description: null,
          isPrivate: true,
        },
      ],
      people: [
        {
          sourceId: "member-1",
          name: "Owner One",
          email: "owner@example.com",
          teamSourceIds: ["board/65"],
        },
        {
          sourceId: "member-2",
          name: "Owner Two",
          email: null,
          teamSourceIds: ["board/65"],
        },
      ],
      labels: [
        {
          sourceId: "label-high",
          name: "P1 High",
          color: "red",
          teamSourceId: "board/65",
        },
      ],
      strategicPillars: [],
      objectives: [],
      keyResults: [],
      sprints: [],
      sourceMetadata: {
        archivedTaskSourceIds: ["65a1234567890abcdef12345"],
        nestedChecklistItemCount: 2,
        platform: "trello",
      },
    });
    expect(draft.warnings).toEqual([
      "1 Trello card comment cannot be imported because comment activity is not supported yet.",
    ]);
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
    expect(draft.sourceMetadata).toBeUndefined();
  });

  it("treats cards in closed Trello lists as archived", () => {
    const draft = createJsonImportDraft({
      fileHash: "archived-list-hash",
      fileName: "archived-board.json",
      text: JSON.stringify({
        cards: [
          {
            closed: false,
            id: "card-on-closed-list",
            idList: "closed-list",
            name: "Archived through its list",
          },
        ],
        id: "archived-board",
        lists: [{ closed: true, id: "closed-list", name: "Archive" }],
        prefs: {},
      }),
    });

    expect(draft.tasks).toHaveLength(1);
    expect(draft.sourceMetadata?.archivedTaskSourceIds).toEqual([
      "card-on-closed-list",
    ]);
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
