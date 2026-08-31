import {
  parseWorkspaceRealtimeEvent,
  WorkspaceRealtimeContractError,
} from "./workspace-realtime-contract";

describe("workspace realtime contract", () => {
  it("decodes supported event shapes", () => {
    expect(parseWorkspaceRealtimeEvent('{"type":"calendar.updated"}')).toEqual({
      kind: "calendar-updated",
    });
    expect(
      parseWorkspaceRealtimeEvent(
        JSON.stringify({
          changes: {
            assigneeId: null,
            autoSchedulingStatus: "scheduled",
            priority: "High",
          },
          storyId: "story-1",
          type: "story.workspace_update",
        }),
      ),
    ).toEqual({
      changes: {
        assigneeId: null,
        autoSchedulingStatus: "scheduled",
        priority: "High",
      },
      kind: "story-updated",
      storyId: "story-1",
    });
    expect(
      parseWorkspaceRealtimeEvent(
        '{"entityId":"story-1","entityType":"story","type":"mention"}',
      ),
    ).toEqual({
      entityId: "story-1",
      entityType: "story",
      kind: "notification",
    });
  });

  it.each([
    "not-json",
    "null",
    "[]",
    "{}",
    '{"type":"story.workspace_update","storyId":"story-1","changes":{}}',
    '{"type":"story.workspace_update","storyId":"story-1","changes":{"priority":"Impossible"}}',
    '{"type":"story.workspace_update","storyId":"story-1","changes":{"unknown":true}}',
  ])("rejects malformed or unsupported payload %s", (payload) => {
    expect(() => parseWorkspaceRealtimeEvent(payload)).toThrow(
      WorkspaceRealtimeContractError,
    );
  });

  it("does not retain malformed event contents in the contract error", () => {
    const payload = '{"providerToken":"secret-value"';

    try {
      parseWorkspaceRealtimeEvent(payload);
      throw new Error("Expected realtime parsing to fail");
    } catch (error) {
      expect(error).toBeInstanceOf(WorkspaceRealtimeContractError);
      expect(String(error)).not.toContain("secret-value");
      expect((error as Error).cause).toBeUndefined();
    }
  });
});
