/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { webcrypto } from "node:crypto";
import { zodTextFormat } from "openai/helpers/zod";
import type { Member } from "@/types";
import type { State } from "@/types/states";
import type { ImportTask, ImportTeam } from "./schema";
import { importAnalysisSchema } from "./schema";
import {
  analyzeImportPersonIdentityClaims,
  deriveImportTeamCode,
  deriveImportTeamColor,
  getAmbiguousImportPersonNameIdentityKeys,
  getBoundedImportSourceKey,
  getImportParentCycleSourceIds,
  getImportPersonIdentityKey,
  getImportPersonSourceIdentityKey,
  isValidImportDate,
  isValidImportDateRange,
  resolveImportAssignee,
  resolveImportEntityByName,
  resolveImportEntityNameMatch,
  resolveImportPerson,
  resolveImportStatus,
  suggestImportPersonMember,
  toImportStoryPayload,
} from "./execution";

const status = (
  id: string,
  name: string,
  category: State["category"],
  isDefault = false,
): State => ({
  id,
  name,
  category,
  isDefault,
  orderIndex: 1,
  color: "#000000",
  teamId: "team",
  workspaceId: "workspace",
  createdAt: "2026-01-01",
  updatedAt: "2026-01-01",
});

const importMember = ({
  email,
  fullName,
  id,
  isActive = true,
  isSystem = false,
  username,
}: {
  email: string;
  fullName: string;
  id: string;
  isActive?: boolean;
  isSystem?: boolean;
  username: string;
}) => ({ email, fullName, id, isActive, isSystem, username }) as Member;

describe("work import execution mapping", () => {
  it("keeps the AI analysis contract convertible to a strict response format", () => {
    Object.defineProperty(globalThis, "structuredClone", {
      configurable: true,
      value: <T>(value: T) => JSON.parse(JSON.stringify(value)) as T,
    });
    expect(() =>
      zodTextFormat(importAnalysisSchema, "work_import_analysis"),
    ).not.toThrow();
  });

  it("prefers an exact status before mapping Jira status categories", () => {
    const statuses = [
      status("default-started", "Started", "started", true),
      status("exact", "In Progress", "started"),
    ];

    expect(resolveImportStatus("Doing now", statuses)?.id).toBe(
      "default-started",
    );
    expect(resolveImportStatus("In Progress", statuses)?.id).toBe("exact");
  });

  it("uses an explicit semantic category for story and objective workflows", () => {
    const storyStatuses = [
      status("todo", "Todo", "unstarted", true),
      status("active", "Active", "started", true),
    ];
    const objectiveStatuses = [
      {
        id: "planned",
        name: "Planned",
        category: "unstarted" as const,
        isDefault: true,
        orderIndex: 1,
      },
      {
        id: "working",
        name: "Working",
        category: "started" as const,
        isDefault: true,
        orderIndex: 2,
      },
    ];

    expect(
      resolveImportStatus("Custom source state", storyStatuses, "started")?.id,
    ).toBe("active");
    expect(
      resolveImportStatus("Custom source state", objectiveStatuses, "started")
        ?.id,
    ).toBe("working");
  });

  it("never chooses an arbitrary duplicate exact status name", () => {
    const statuses = [
      status("todo", "Todo", "unstarted", true),
      status("started-default", "Working", "started", true),
      status("started-secondary", "Working", "started"),
      status("completed", "Working", "completed", true),
    ];
    statuses[1].orderIndex = 1;
    statuses[2].orderIndex = 2;

    expect(resolveImportStatus("Working", statuses, "completed")?.id).toBe(
      "completed",
    );
    expect(resolveImportStatus("Working", statuses, "started")?.id).toBe(
      "started-default",
    );
    expect(resolveImportStatus("Working", statuses)?.id).toBe("todo");
  });

  it("maps completed Jira statuses and falls back to the default unstarted state", () => {
    const statuses = [
      status("todo", "Todo", "unstarted", true),
      status("done", "Finished", "completed", true),
    ];

    expect(resolveImportStatus("Resolved", statuses)?.id).toBe("done");
    expect(resolveImportStatus("Unknown state", statuses)?.id).toBe("todo");
  });

  it("only assigns an active team member with an exact email match", () => {
    const member = importMember({
      email: "Owner@Example.com",
      fullName: "Owner Name",
      id: "member-1",
      username: "owner",
    });

    expect(resolveImportAssignee("owner@example.com", [member])?.id).toBe(
      "member-1",
    );
    expect(
      resolveImportAssignee("different@example.com", [member]),
    ).toBeUndefined();
  });

  it("prefers exact email and reports unique name matches for review", () => {
    const emailMember = importMember({
      email: "ada@example.com",
      fullName: "Different Name",
      id: "member-email",
      username: "ada-email",
    });
    const nameMember = importMember({
      email: "someone@example.com",
      fullName: "Ada Lovelace",
      id: "member-name",
      username: "ada",
    });

    expect(
      resolveImportPerson(
        { email: " ADA@example.com ", name: "Ada Lovelace" },
        [emailMember, nameMember],
      ),
    ).toEqual({
      matchedBy: "email",
      member: emailMember,
      requiresReview: false,
    });

    expect(
      resolveImportPerson(
        { email: "external@example.com", name: "Ada Lovelace" },
        [nameMember],
      ),
    ).toBeUndefined();
    expect(
      resolveImportPerson({ email: null, name: "  ada   lovelace " }, [
        emailMember,
        nameMember,
      ]),
    ).toEqual({
      matchedBy: "fullName",
      member: nameMember,
      requiresReview: true,
    });
    expect(
      resolveImportPerson({ email: null, name: "ADA" }, [
        emailMember,
        nameMember,
      ])?.matchedBy,
    ).toBe("username");
    expect(
      resolveImportAssignee(null, [emailMember, nameMember], "Ada Lovelace"),
    ).toBeUndefined();
  });

  it("does not resolve ambiguous, inactive, or system member names", () => {
    const members = [
      importMember({
        email: "one@example.com",
        fullName: "Alex Kim",
        id: "one",
        username: "alex-one",
      }),
      importMember({
        email: "two@example.com",
        fullName: "Alex Kim",
        id: "two",
        username: "alex-two",
      }),
      importMember({
        email: "inactive@example.com",
        fullName: "Inactive Person",
        id: "inactive",
        isActive: false,
        username: "inactive",
      }),
      importMember({
        email: "system@example.com",
        fullName: "System Person",
        id: "system",
        isSystem: true,
        username: "system",
      }),
    ];

    expect(
      resolveImportPerson({ email: null, name: "Alex Kim" }, members),
    ).toBeUndefined();
    expect(
      resolveImportPerson({ email: null, name: "Inactive Person" }, members),
    ).toBeUndefined();
    expect(
      resolveImportPerson({ email: null, name: "System Person" }, members),
    ).toBeUndefined();
    expect(
      resolveImportPerson({ email: null, name: null }, members),
    ).toBeUndefined();
  });

  it("suggests one close member name without treating it as an assignment", () => {
    const likelyMember = importMember({
      email: "joseph@example.com",
      fullName: "Joseph Smith",
      id: "joseph",
      username: "joseph-smith",
    });
    const unrelatedMember = importMember({
      email: "mary@example.com",
      fullName: "Mary Jones",
      id: "mary",
      username: "mary",
    });

    expect(
      suggestImportPersonMember({ name: "Joe Smith" }, [
        unrelatedMember,
        likelyMember,
      ])?.member.id,
    ).toBe("joseph");
    expect(
      suggestImportPersonMember({ name: "Alex Kim" }, [
        importMember({
          email: "alex-one@example.com",
          fullName: "Alex Kim",
          id: "alex-one",
          username: "alex-one",
        }),
        importMember({
          email: "alex-two@example.com",
          fullName: "Alex Kim",
          id: "alex-two",
          username: "alex-two",
        }),
      ]),
    ).toBeUndefined();
  });

  it("normalizes unresolved people by email, name, then stable source identity", () => {
    expect(
      getImportPersonIdentityKey({
        email: " OWNER@Example.com ",
        name: "Different Name",
      }),
    ).toBe("email:owner@example.com");
    expect(
      getImportPersonIdentityKey({ email: null, name: "  Ada   Lovelace " }),
    ).toBe("name:ada lovelace");
    expect(getImportPersonIdentityKey(undefined, " Person-1 ")).toBe(
      "source:Person-1",
    );
  });

  it("keeps opaque source IDs case-sensitive and detects conflicting claims", () => {
    const identities = [
      {
        identity: { email: "one@example.com", name: "One Person" },
        sourceId: "Person-1",
      },
      {
        identity: { email: "two@example.com", name: "Two Person" },
        sourceId: "Person-1",
      },
      {
        identity: { email: "case@example.com", name: "Case Person" },
        sourceId: "person-1",
      },
    ];
    const analysis = analyzeImportPersonIdentityClaims(identities);

    expect(
      getImportPersonSourceIdentityKey(identities[0].identity, "Person-1"),
    ).toBe("source:Person-1");
    expect(
      getImportPersonSourceIdentityKey(identities[2].identity, "person-1"),
    ).toBe("source:person-1");
    expect(analysis.conflictedIdentityKeys).toEqual(
      new Set(["source:Person-1"]),
    );
    expect(analysis.canonicalIdentities.get("source:person-1")).toEqual({
      email: "case@example.com",
      name: "Case Person",
    });
  });

  it("keeps distinct source people separate and detects duplicate source names", () => {
    const identities = [
      {
        identity: { email: "alex@example.com", name: "Alex Kim" },
        sourceId: "person-1",
      },
      {
        identity: { email: null, name: " Alex  Kim " },
        sourceId: "person-2",
      },
      {
        identity: { email: null, name: "Alex Kim" },
        sourceId: "person-2",
      },
    ];

    expect(
      getImportPersonSourceIdentityKey(identities[0].identity, "person-1"),
    ).toBe("source:person-1");
    expect(
      [...getAmbiguousImportPersonNameIdentityKeys(identities)].sort(),
    ).toEqual(["source:person-1", "source:person-2"]);
    expect(
      getAmbiguousImportPersonNameIdentityKeys([identities[1], identities[2]]),
    ).toEqual(new Set());
  });

  it("only reuses a unique exact normalized entity name in the destination team", () => {
    const entities = [
      { id: "team-1-roadmap", name: "Product   Roadmap", teamId: "team-1" },
      { id: "team-2-roadmap", name: "Product Roadmap", teamId: "team-2" },
    ];

    expect(
      resolveImportEntityByName(" product roadmap ", "team-1", entities)?.id,
    ).toBe("team-1-roadmap");
    expect(
      resolveImportEntityByName("Product Roadmap", "team-3", entities),
    ).toBeUndefined();
    expect(
      resolveImportEntityByName("Product Roadmap", "team-1", [
        ...entities,
        { id: "duplicate", name: "product roadmap", teamId: "team-1" },
      ]),
    ).toBeUndefined();
    expect(
      resolveImportEntityNameMatch("Product Roadmap", [
        entities[0],
        { id: "duplicate", name: "product roadmap", teamId: "team-1" },
      ]),
    ).toMatchObject({ kind: "ambiguous" });
  });

  it("validates real calendar dates and ordered optional date ranges", () => {
    expect(isValidImportDate("2028-02-29")).toBe(true);
    expect(isValidImportDate("2027-02-29")).toBe(false);
    expect(isValidImportDate("2026-13-01")).toBe(false);
    expect(isValidImportDate("01/02/2026")).toBe(false);
    expect(isValidImportDateRange("", null)).toBe(false);
    expect(isValidImportDateRange(null, null)).toBe(true);
    expect(isValidImportDateRange("2026-01-01", "2026-01-31")).toBe(true);
    expect(isValidImportDateRange("2026-02-01", "2026-01-31")).toBe(false);
  });

  it("finds only selected, same-destination stories inside parent cycles", () => {
    const task = (
      sourceId: string,
      parentSourceId: string | null,
      teamSourceId = "team-1",
    ) => ({ parentSourceId, sourceId, teamSourceId }) as ImportTask;
    const tasks = [
      task("A", "B"),
      task("B", "A"),
      task("child", "A"),
      task("cross-1", "cross-2", "team-1"),
      task("cross-2", "cross-1", "team-2"),
      task("duplicate", null),
      task("duplicate", null),
    ];

    expect(
      [
        ...getImportParentCycleSourceIds(
          tasks,
          (child, parent) => child.teamSourceId === parent.teamSourceId,
        ),
      ].sort(),
    ).toEqual(["A", "B"]);
  });

  it("derives deterministic, bounded team codes and colors", () => {
    const team = {
      sourceId: "source-team-1",
      name: "Platform Engineering",
      code: null,
      color: null,
    } as ImportTeam;

    expect(deriveImportTeamCode(team)).toBe("PE");
    const alternative = deriveImportTeamCode(team, ["PE"]);
    expect(alternative).toMatch(/^PE[A-Z0-9]$/);
    expect(deriveImportTeamCode(team, ["PE"])).toBe(alternative);
    expect(deriveImportTeamColor(team)).toMatch(/^#[0-9A-F]{6}$/);
    expect(deriveImportTeamColor(team)).toBe(deriveImportTeamColor(team));
    expect(deriveImportTeamColor({ ...team, color: "#aabbcc" })).toBe(
      "#AABBCC",
    );
  });

  it("builds a story payload with resolved multi-entity references", () => {
    const task = {
      sourceId: "task-1",
      title: "  Ship migration  ",
      description: "  Complete the migration safely.  ",
      status: "Doing",
      statusCategory: "started",
      priority: "High",
      estimateValue: 5,
      estimatedDurationMinutes: 90,
      minimumFocusBlockMinutes: 30,
      assigneeEmail: null,
      assigneeName: "Ada Lovelace",
      startDate: "2026-01-01",
      endDate: "2026-01-31",
    } as ImportTask;
    const member = importMember({
      email: "ada@example.com",
      fullName: "Ada Lovelace",
      id: "member-1",
      username: "ada",
    });

    expect(
      toImportStoryPayload({
        assigneeId: "member-1",
        keyResultId: "key-result-1",
        labelIds: ["label-1", "label-1", "label-2"],
        members: [member],
        objectiveId: "objective-1",
        parentId: "parent-1",
        sprintId: "sprint-1",
        statuses: [status("started", "In progress", "started", true)],
        task,
        teamId: "team-1",
      }),
    ).toEqual({
      assigneeId: "member-1",
      description: "Complete the migration safely.",
      endDate: "2026-01-31",
      keyResultId: "key-result-1",
      labelIds: ["label-1", "label-2"],
      estimateValue: 5,
      estimatedDurationMinutes: 90,
      minimumFocusBlockMinutes: 30,
      objectiveId: "objective-1",
      parentId: "parent-1",
      priority: "High",
      sprintId: "sprint-1",
      startDate: "2026-01-01",
      statusId: "started",
      teamId: "team-1",
      title: "Ship migration",
    });
  });

  it("includes rich HTML for imported Markdown descriptions", () => {
    const task = {
      sourceId: "task-with-markdown",
      title: "Imported checklist",
      description: "### Checklist\n\n- [ ] Upload proof\n- [x] Link proof",
      status: null,
      statusCategory: null,
      priority: "No Priority",
      assigneeEmail: null,
      assigneeName: null,
      startDate: null,
      endDate: null,
    } as ImportTask;

    expect(
      toImportStoryPayload({
        members: [],
        statuses: [],
        task,
        teamId: "team-1",
      }),
    ).toMatchObject({
      description: task.description,
      descriptionHTML:
        '<h3>Checklist</h3><ul data-type="taskList"><li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label><div><p>Upload proof</p></div></li><li data-checked="true" data-type="taskItem"><label><input type="checkbox" checked="checked"><span></span></label><div><p>Link proof</p></div></li></ul>',
    });
  });

  it("rejects an invalid task date range before constructing a payload", () => {
    const task = {
      sourceId: "task-1",
      title: "Bad dates",
      description: "",
      status: null,
      statusCategory: null,
      priority: "No Priority",
      assigneeEmail: null,
      assigneeName: null,
      startDate: "2026-02-01",
      endDate: "2026-01-01",
    } as ImportTask;

    expect(() =>
      toImportStoryPayload({
        members: [],
        statuses: [],
        task,
        teamId: "team-1",
      }),
    ).toThrow("Imported task task-1 has an invalid date range");
  });

  it("rejects effort values that violate the story domain contract", () => {
    const invalidTasks = [
      {
        ...({
          sourceId: "invalid-estimate",
          title: "Invalid estimate",
          description: "",
          status: null,
          statusCategory: null,
          priority: "No Priority",
          assigneeEmail: null,
          assigneeName: null,
          startDate: null,
          endDate: null,
        } as ImportTask),
        estimateValue: 4,
        estimatedDurationMinutes: null,
        minimumFocusBlockMinutes: null,
      },
      {
        ...({
          sourceId: "missing-duration",
          title: "Missing duration",
          description: "",
          status: null,
          statusCategory: null,
          priority: "No Priority",
          assigneeEmail: null,
          assigneeName: null,
          startDate: null,
          endDate: null,
        } as ImportTask),
        estimateValue: null,
        estimatedDurationMinutes: null,
        minimumFocusBlockMinutes: 30,
      },
    ];

    for (const task of invalidTasks) {
      expect(() =>
        toImportStoryPayload({
          members: [],
          statuses: [],
          task: task as ImportTask,
          teamId: "team-1",
        }),
      ).toThrow(`Imported task ${task.sourceId} has invalid effort values`);
    }
  });

  it("keeps ordinary Jira keys readable and hashes unsafe or oversized source keys", async () => {
    Object.defineProperty(globalThis.crypto, "subtle", {
      configurable: true,
      value: webcrypto.subtle,
    });

    await expect(getBoundedImportSourceKey(" JIRA-42 ")).resolves.toBe(
      "JIRA-42",
    );

    const unsafe = await getBoundedImportSourceKey("row\n42");
    const oversized = await getBoundedImportSourceKey("🚀".repeat(100));

    expect(unsafe).toMatch(/^source:[a-f0-9]{64}$/);
    expect(oversized).toMatch(/^source:[a-f0-9]{64}$/);
    await expect(getBoundedImportSourceKey("row\n42")).resolves.toBe(unsafe);
  });
});
