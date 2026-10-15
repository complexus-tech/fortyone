import { describe, expect, it } from "vitest";
import { findStories, formatTime, resourceBaseVersion } from "./model";
import type { WorkspaceData } from "./types";

const data: WorkspaceData = {
  workspaceId: "workspace",
  teams: [{ id: "team", name: "Engineering", code: "ENG" }],
  statuses: [],
  comments: [],
  documents: [],
  pending: [],
  lastSyncedAt: null,
  stories: [
    {
      id: "1",
      title: "Prepare launch",
      description: "Check installer",
      statusId: "open",
      teamId: "team",
      assigneeId: "user",
      sequenceId: 42,
      priority: "High",
      updatedAt: "2026-10-14",
    },
    {
      id: "2",
      title: "Write release notes",
      statusId: "open",
      teamId: "team",
      assigneeId: "another-user",
      sequenceId: 43,
      priority: "Low",
      updatedAt: "2026-10-14",
    },
  ],
};

describe("offline story search", () => {
  it("limits My work to the signed-in user", () =>
    expect(
      findStories(data, "user", "", true).map((story) => story.id),
    ).toEqual(["1"]));
  it("searches cached descriptions and references without a server request", () => {
    expect(findStories(data, "user", " INSTALLER ", false)[0]?.id).toBe("1");
    expect(findStories(data, "user", "eng-43", false)[0]?.id).toBe("2");
  });
  it("searches the canonical server reference when sequenceId is not present", () => {
    const synced = {
      ...data,
      stories: [
        { ...data.stories[0]!, reference: "ENG-123", sequenceId: undefined },
      ],
    };
    expect(findStories(synced, "user", "ENG-123", false)).toHaveLength(1);
  });
  it("never labels an invalid sync timestamp as successful", () => {
    expect(formatTime(null)).toBe("Not downloaded yet");
    expect(formatTime("invalid")).toBe("Unknown");
  });
});

describe("versions for offline edits", () => {
  it("uses a dependency marker only for a matching unsynced create", () => {
    const pending = [
      {
        id: "op",
        kind: "story.create",
        resourceId: "story",
        payload: {},
        baseVersion: null,
        status: "pending" as const,
        error: null,
      },
    ];
    expect(resourceBaseVersion("story", "story", undefined, pending)).toBe(
      "pending-create:story",
    );
    expect(
      resourceBaseVersion("document", "story", undefined, pending),
    ).toBeNull();
    expect(
      resourceBaseVersion("story", "other", undefined, pending),
    ).toBeNull();
    expect(resourceBaseVersion("story", "story", "server-time", pending)).toBe(
      "server-time",
    );
    expect(resourceBaseVersion("document", "document", 0, [])).toBe("0");
  });
});
