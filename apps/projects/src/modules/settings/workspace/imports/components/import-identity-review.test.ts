/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { Member } from "@/types";
import type { ImportTask } from "../schema";
import { resolveImportPerson } from "../execution";
import { collectImportIdentities } from "./import-identity-review";

const task = (overrides: Partial<ImportTask> = {}): ImportTask => ({
  assigneeEmail: null,
  assigneeName: null,
  assigneePersonSourceId: null,
  associations: [],
  collaboratorPersonSourceIds: [],
  description: "",
  endDate: null,
  estimateValue: null,
  estimatedDurationMinutes: null,
  keyResultSourceId: null,
  labelSourceIds: [],
  links: [],
  minimumFocusBlockMinutes: null,
  objectiveSourceId: null,
  parentSourceId: null,
  priority: "No Priority",
  sourceId: "task-1",
  sprintSourceId: null,
  startDate: null,
  status: null,
  statusCategory: null,
  teamSourceId: null,
  title: "Imported work",
  ...overrides,
});

const member = (overrides: Partial<Member>): Member => ({
  avatarUrl: null,
  createdAt: "2026-01-01T00:00:00.000Z",
  email: "member@example.com",
  fullName: "Workspace Member",
  id: "member-1",
  isActive: true,
  isInternal: false,
  isSystem: false,
  role: "member",
  updatedAt: "2026-01-01T00:00:00.000Z",
  username: "member",
  ...overrides,
});

describe("import identity review", () => {
  it("uses a later exact email claim before suggesting by the first name", () => {
    const [identity] = collectImportIdentities({
      keyResults: [],
      objectives: [],
      people: [
        {
          email: null,
          name: "Alex Kim",
          sourceId: "person-1",
          teamSourceIds: [],
        },
      ],
      tasks: [
        task({
          assigneeEmail: "owner@example.com",
          assigneePersonSourceId: "person-1",
        }),
      ],
    });
    const resolution = resolveImportPerson(identity, [
      member({
        email: "different@example.com",
        fullName: "Alex Kim",
        id: "name-match",
        username: "alex",
      }),
      member({
        email: "owner@example.com",
        fullName: "Different Name",
        id: "email-match",
        username: "owner",
      }),
    ]);

    expect(identity).toMatchObject({
      email: "owner@example.com",
      hasConflictingClaims: false,
      identityKey: "source:person-1",
      name: "Alex Kim",
    });
    expect(resolution?.member.id).toBe("email-match");
    expect(resolution?.matchedBy).toBe("email");
  });

  it("does not choose a default when one source ID has conflicting emails", () => {
    const [identity] = collectImportIdentities({
      keyResults: [],
      objectives: [],
      people: [
        {
          email: null,
          name: "Alex Kim",
          sourceId: "person-1",
          teamSourceIds: [],
        },
      ],
      tasks: [
        task({
          assigneeEmail: "one@example.com",
          assigneePersonSourceId: "person-1",
          sourceId: "task-1",
        }),
        task({
          assigneeEmail: "two@example.com",
          assigneePersonSourceId: "person-1",
          sourceId: "task-2",
        }),
      ],
    });

    expect(identity).toMatchObject({
      email: null,
      hasConflictingClaims: true,
      identityKey: "source:person-1",
    });
  });
});
