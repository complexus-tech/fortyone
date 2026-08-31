/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { createJsonImportDraft } from "./json";

describe("work import JSON parsing", () => {
  it("creates a vendor-neutral preview from a nested board export", () => {
    const draft = createJsonImportDraft({
      fileHash: "trello-hash",
      fileName: "product-board.json",
      text: JSON.stringify({
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
      }),
    });

    expect(draft.sourceType).toBe("json");
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
    expect(draft.tasks[0]).toEqual(
      expect.objectContaining({
        priority: "Urgent",
        sourceId: "task-1",
        title: "Move the backlog",
      }),
    );
    expect(draft.rows[0]?.metadata).toBe('{"source":"legacy"}');
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
});
