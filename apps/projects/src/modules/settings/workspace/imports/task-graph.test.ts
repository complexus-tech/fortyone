/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { ImportTask } from "./schema";
import { mergeAnalyzedTaskGraph } from "./task-graph";

const task = (
  sourceId: string,
  title: string,
  overrides: Partial<ImportTask> = {},
): ImportTask => ({
  sourceId,
  title,
  description: "",
  status: null,
  statusCategory: null,
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
  ...overrides,
});

describe("analyzed task graph merge", () => {
  it("prefers an exact source ID over the title fallback", () => {
    const [merged] = mergeAnalyzedTaskGraph(
      [task("row-1", "Deterministic title")],
      [
        task("row-1", "AI renamed title", { teamSourceId: "exact-team" }),
        task("other", "Deterministic title", {
          teamSourceId: "title-team",
        }),
      ],
    );

    expect(merged.teamSourceId).toBe("exact-team");
    expect(merged.title).toBe("Deterministic title");
  });

  it("merges collaborator relationships without replacing source task content", () => {
    const [merged] = mergeAnalyzedTaskGraph(
      [task("row-1", "Keep deterministic content")],
      [
        task("row-1", "AI title", {
          collaboratorPersonSourceIds: ["person-2", "person-3"],
        }),
      ],
    );

    expect(merged.collaboratorPersonSourceIds).toEqual([
      "person-2",
      "person-3",
    ]);
    expect(merged.title).toBe("Keep deterministic content");
  });

  it("treats a sparse AI result as enrichment without erasing deterministic source values", () => {
    const [merged] = mergeAnalyzedTaskGraph(
      [
        task("card-1", "Keep deterministic content", {
          assigneeName: "Source owner",
          assigneePersonSourceId: "person-1",
          collaboratorPersonSourceIds: ["person-2"],
          labelSourceIds: ["label-1"],
          links: [{ title: "Trello card", url: "https://trello.com/c/card-1" }],
          teamSourceId: "team-1",
        }),
      ],
      [
        task("card-1", "Schema-required title", {
          objectiveSourceId: "objective-1",
          priority: "High",
        }),
      ],
      { enrichmentOnly: true },
    );

    expect(merged).toMatchObject({
      assigneeName: "Source owner",
      assigneePersonSourceId: "person-1",
      collaboratorPersonSourceIds: ["person-2"],
      labelSourceIds: ["label-1"],
      links: [{ title: "Trello card", url: "https://trello.com/c/card-1" }],
      objectiveSourceId: "objective-1",
      priority: "High",
      teamSourceId: "team-1",
      title: "Keep deterministic content",
    });
  });

  it("requires an exact source ID for enrichment-only task patches", () => {
    const [merged] = mergeAnalyzedTaskGraph(
      [task("card-1", "Same title")],
      [task("hallucinated-card", "Same title", { priority: "Urgent" })],
      { enrichmentOnly: true },
    );

    expect(merged.priority).toBe("No Priority");
  });

  it("merges only explicit analyzed story effort values", () => {
    const [merged] = mergeAnalyzedTaskGraph(
      [task("row-1", "Keep deterministic content")],
      [
        task("row-1", "AI title", {
          estimateValue: 5,
          estimatedDurationMinutes: 90,
          minimumFocusBlockMinutes: 30,
        }),
      ],
    );

    expect(merged).toMatchObject({
      estimateValue: 5,
      estimatedDurationMinutes: 90,
      minimumFocusBlockMinutes: 30,
    });
  });

  it("merges explicit story associations from the analyzed graph", () => {
    const [merged] = mergeAnalyzedTaskGraph(
      [task("row-1", "Keep deterministic content")],
      [
        task("row-1", "AI title", {
          associations: [{ type: "blocked_by", targetSourceId: "story-2" }],
        }),
      ],
    );

    expect(merged.associations).toEqual([
      { type: "blocked_by", targetSourceId: "story-2" },
    ]);
    expect(merged.title).toBe("Keep deterministic content");
  });

  it("merges, canonicalizes, and deterministically deduplicates safe story links", () => {
    const [merged] = mergeAnalyzedTaskGraph(
      [task("row-1", "Keep deterministic content")],
      [
        task("row-1", "AI title", {
          links: [
            {
              title: " Original card ",
              url: "HTTPS://EXAMPLE.COM:443/card/1",
            },
            {
              title: "Duplicate card",
              url: "https://example.com/card/1",
            },
            { title: "Attachment", url: "https://cdn.example.com/file.pdf" },
          ],
        }),
      ],
    );

    expect(merged.links).toEqual([
      { title: "Original card", url: "https://example.com/card/1" },
      { title: "Attachment", url: "https://cdn.example.com/file.pdf" },
    ]);
  });

  it("keeps only an edited priority authoritative", () => {
    const [merged] = mergeAnalyzedTaskGraph(
      [
        task("row-1", "Reviewed title", {
          priority: "Low",
          status: "Ready",
        }),
      ],
      [
        task("row-1", "AI title", {
          assigneeName: "AI assignee",
          assigneePersonSourceId: "person-1",
          objectiveSourceId: "objective-1",
          priority: "Urgent",
          statusCategory: "started",
          teamSourceId: "team-1",
        }),
      ],
      { authoritativeFields: new Set(["priority"]) },
    );

    expect(merged).toMatchObject({
      assigneeName: "AI assignee",
      assigneePersonSourceId: "person-1",
      objectiveSourceId: "objective-1",
      priority: "Low",
      status: "Ready",
      statusCategory: "started",
      teamSourceId: "team-1",
    });
  });

  it("keeps only an edited status authoritative", () => {
    const [merged] = mergeAnalyzedTaskGraph(
      [task("row-1", "Reviewed title", { status: "Ready" })],
      [
        task("row-1", "AI title", {
          assigneeName: "AI assignee",
          assigneePersonSourceId: "person-1",
          priority: "High",
          statusCategory: "completed",
        }),
      ],
      { authoritativeFields: new Set(["status"]) },
    );

    expect(merged).toMatchObject({
      assigneeName: "AI assignee",
      assigneePersonSourceId: "person-1",
      priority: "High",
      status: "Ready",
      statusCategory: null,
    });
  });

  it("keeps only an edited assignee column authoritative", () => {
    const [merged] = mergeAnalyzedTaskGraph(
      [
        task("row-1", "Reviewed title", {
          assigneeEmail: "reviewed@example.com",
        }),
      ],
      [
        task("row-1", "AI title", {
          assigneeName: "Stale AI assignee",
          assigneePersonSourceId: "person-stale",
          priority: "High",
          statusCategory: "completed",
        }),
      ],
      { authoritativeFields: new Set(["assigneeEmail"]) },
    );

    expect(merged).toMatchObject({
      assigneeEmail: "reviewed@example.com",
      assigneeName: null,
      assigneePersonSourceId: null,
      priority: "High",
      statusCategory: "completed",
    });
  });

  it("falls back to one normalized title match when source IDs differ", () => {
    const [merged] = mergeAnalyzedTaskGraph(
      [task("row-1", "  Ship   Migration ")],
      [
        task("ai-task-9", "ship migration", {
          objectiveSourceId: "objective-1",
          priority: "Urgent",
        }),
      ],
    );

    expect(merged.objectiveSourceId).toBe("objective-1");
    expect(merged.priority).toBe("Urgent");
  });

  it("does not enrich ambiguous duplicate source IDs or titles", () => {
    const original = task("row-1", "Duplicate task");
    const [merged] = mergeAnalyzedTaskGraph(
      [original],
      [
        task("row-1", "Duplicate task", { teamSourceId: "team-1" }),
        task("row-1", "duplicate task", { teamSourceId: "team-2" }),
      ],
    );

    expect(merged).toEqual(original);
  });

  it("does not apply one AI title match to multiple deterministic rows", () => {
    const originalTasks = [
      task("row-1", "Repeated title"),
      task("row-2", "Repeated title"),
    ];
    const merged = mergeAnalyzedTaskGraph(originalTasks, [
      task("different", "Repeated title", { teamSourceId: "team-1" }),
    ]);

    expect(merged).toEqual(originalTasks);
  });

  it("drops AI task-to-task references when source IDs were remapped", () => {
    const [merged] = mergeAnalyzedTaskGraph(
      [task("new-1", "Remapped task")],
      [
        task("old-1", "Remapped task", {
          parentSourceId: "old-parent",
          associations: [{ type: "related", targetSourceId: "old-2" }],
          objectiveSourceId: "objective-1",
        }),
      ],
      { authoritativeFields: new Set(["sourceId"]) },
    );

    expect(merged.parentSourceId).toBeNull();
    expect(merged.associations).toEqual([]);
    expect(merged.objectiveSourceId).toBe("objective-1");
  });
});
