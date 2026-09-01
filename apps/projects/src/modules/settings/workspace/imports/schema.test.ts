/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { importAnalysisSchema } from "./schema";

const validAnalysis = {
  sourceType: "json",
  sourceNamespace: "trello:board:board-1",
  summary: "Found a structured migration plan.",
  warnings: ["One unresolved member needs review."],
  mapping: null,
  teams: [
    {
      sourceId: "team-1",
      name: "Product",
      code: "PROD",
      color: "#3366FF",
      description: "Product delivery",
      isPrivate: false,
    },
  ],
  people: [
    {
      sourceId: "person-1",
      name: "Owner",
      email: "owner@example.com",
      teamSourceIds: ["team-1"],
    },
  ],
  labels: [
    {
      sourceId: "label-1",
      name: "Migration",
      color: "blue",
      teamSourceId: "team-1",
    },
  ],
  strategicPillars: [
    {
      sourceId: "pillar-1",
      name: "Customer growth",
      description: "Grow the active customer base",
      orderIndex: 0,
    },
  ],
  objectives: [
    {
      sourceId: "objective-1",
      name: "Complete the migration",
      description: null,
      shortSummary: "Move active work safely",
      color: "#3366FF",
      isPrivate: false,
      status: "In progress",
      statusCategory: "started",
      priority: "High",
      leadPersonSourceId: "person-1",
      teamSourceId: "team-1",
      pillarSourceId: "pillar-1",
      startDate: "2026-09-01",
      endDate: "2026-09-30",
    },
  ],
  keyResults: [
    {
      sourceId: "kr-1",
      name: "Move every active card",
      objectiveSourceId: "objective-1",
      measurementType: "percentage",
      startValue: 0,
      currentValue: 25,
      targetValue: 100,
      leadPersonSourceId: "person-1",
      contributorPersonSourceIds: ["person-1"],
      startDate: "2026-09-01",
      endDate: "2026-09-30",
    },
  ],
  sprints: [
    {
      sourceId: "sprint-1",
      name: "Migration week",
      goal: "Move the active backlog",
      teamSourceId: "team-1",
      objectiveSourceId: "objective-1",
      startDate: "2026-09-01",
      endDate: "2026-09-07",
    },
  ],
  tasks: [
    {
      sourceId: "story-1",
      title: "Move the backlog",
      description: "Migrate active work.",
      status: "Doing",
      statusCategory: "started",
      priority: "High",
      estimateValue: 5,
      estimatedDurationMinutes: 90,
      minimumFocusBlockMinutes: 30,
      assigneeEmail: "owner@example.com",
      assigneeName: "Owner",
      assigneePersonSourceId: "person-1",
      collaboratorPersonSourceIds: [],
      teamSourceId: "team-1",
      parentSourceId: null,
      objectiveSourceId: "objective-1",
      keyResultSourceId: "kr-1",
      sprintSourceId: "sprint-1",
      labelSourceIds: ["label-1"],
      associations: [{ type: "blocked_by", targetSourceId: "story-2" }],
      links: [
        {
          title: "Original Trello card",
          url: "https://trello.com/c/example",
        },
      ],
      startDate: "2026-09-01",
      endDate: "2026-09-05",
    },
  ],
} as const;

describe("work import analysis schema", () => {
  it("accepts a strict vendor-neutral entity graph with source relationships", () => {
    expect(importAnalysisSchema.parse(validAnalysis)).toMatchObject({
      teams: [{ sourceId: "team-1" }],
      strategicPillars: [{ sourceId: "pillar-1" }],
      objectives: [
        { leadPersonSourceId: "person-1", pillarSourceId: "pillar-1" },
      ],
      keyResults: [{ objectiveSourceId: "objective-1" }],
      tasks: [
        {
          teamSourceId: "team-1",
          objectiveSourceId: "objective-1",
          collaboratorPersonSourceIds: [],
          associations: [{ type: "blocked_by", targetSourceId: "story-2" }],
          links: [
            {
              title: "Original Trello card",
              url: "https://trello.com/c/example",
            },
          ],
          labelSourceIds: ["label-1"],
        },
      ],
    });
  });

  it("rejects undeclared fields instead of accepting mutation instructions", () => {
    const result = importAnalysisSchema.safeParse({
      ...validAnalysis,
      teams: [{ ...validAnalysis.teams[0], destinationTeamId: "internal-id" }],
    });

    expect(result.success).toBe(false);
  });

  it("requires every entity collection so AI output cannot omit part of the contract", () => {
    const { people: _people, ...missingPeople } = validAnalysis;

    expect(importAnalysisSchema.safeParse(missingPeople).success).toBe(false);
  });

  it("bounds sprint goals to the import and creation contract", () => {
    const result = importAnalysisSchema.safeParse({
      ...validAnalysis,
      sprints: [{ ...validAnalysis.sprints[0], goal: "x".repeat(10_001) }],
    });

    expect(result.success).toBe(false);
  });

  it("bounds strategic pillar ordering to a nonnegative PostgreSQL int4", () => {
    for (const orderIndex of [-1, 1.5, 2_147_483_648]) {
      const result = importAnalysisSchema.safeParse({
        ...validAnalysis,
        strategicPillars: [
          { ...validAnalysis.strategicPillars[0], orderIndex },
        ],
      });

      expect(result.success).toBe(false);
    }
  });

  it("rejects undeclared or unsupported task association fields", () => {
    const result = importAnalysisSchema.safeParse({
      ...validAnalysis,
      tasks: [
        {
          ...validAnalysis.tasks[0],
          associations: [
            { type: "depends_on", targetSourceId: "story-2", weight: 2 },
          ],
        },
      ],
    });

    expect(result.success).toBe(false);
  });

  it("accepts only bounded absolute HTTP or HTTPS story links", () => {
    for (const url of [
      "/relative/card-1",
      "data:text/plain,unsafe",
      "file:///tmp/export.json",
      ["javascript", "alert(1)"].join(":"),
      "https://user:secret@example.com/card-1",
      `https://example.com/${"x".repeat(256)}`,
    ]) {
      const result = importAnalysisSchema.safeParse({
        ...validAnalysis,
        tasks: [
          {
            ...validAnalysis.tasks[0],
            links: [{ title: null, url }],
          },
        ],
      });

      expect(result.success).toBe(false);
    }

    for (const links of [
      [
        {
          title: "x".repeat(256),
          url: "https://example.com/card-1",
        },
      ],
      Array.from({ length: 101 }, (_, index) => ({
        title: null,
        url: `https://example.com/card-${index}`,
      })),
    ]) {
      const result = importAnalysisSchema.safeParse({
        ...validAnalysis,
        tasks: [{ ...validAnalysis.tasks[0], links }],
      });

      expect(result.success).toBe(false);
    }

    const undeclaredField = importAnalysisSchema.safeParse({
      ...validAnalysis,
      tasks: [
        {
          ...validAnalysis.tasks[0],
          links: [
            {
              title: null,
              url: "https://example.com/card-1",
              download: true,
            },
          ],
        },
      ],
    });
    expect(undeclaredField.success).toBe(false);
  });

  it("accepts only backend-supported story effort values", () => {
    for (const estimateValue of [0, 4, 13, 1.5]) {
      const result = importAnalysisSchema.safeParse({
        ...validAnalysis,
        tasks: [{ ...validAnalysis.tasks[0], estimateValue }],
      });

      expect(result.success).toBe(false);
    }

    for (const estimatedDurationMinutes of [0, 1.5, 2_401]) {
      const result = importAnalysisSchema.safeParse({
        ...validAnalysis,
        tasks: [{ ...validAnalysis.tasks[0], estimatedDurationMinutes }],
      });

      expect(result.success).toBe(false);
    }
  });

  it("requires minimum focus blocks to fit inside an estimated duration", () => {
    expect(
      importAnalysisSchema.safeParse({
        ...validAnalysis,
        tasks: [
          {
            ...validAnalysis.tasks[0],
            estimatedDurationMinutes: null,
            minimumFocusBlockMinutes: 30,
          },
        ],
      }).success,
    ).toBe(false);
    expect(
      importAnalysisSchema.safeParse({
        ...validAnalysis,
        tasks: [
          {
            ...validAnalysis.tasks[0],
            estimatedDurationMinutes: 30,
            minimumFocusBlockMinutes: 60,
          },
        ],
      }).success,
    ).toBe(false);
  });
});
