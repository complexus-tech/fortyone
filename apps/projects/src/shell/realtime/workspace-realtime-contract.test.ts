import {
  parseWorkspaceRealtimeEvent,
  WorkspaceRealtimeContractError,
} from "./workspace-realtime-contract";

describe("workspace realtime contract", () => {
  it("decodes the server's snake-case notification stream", () => {
    expect(
      parseWorkspaceRealtimeEvent(
        '{"entity_id":"story-1","entity_type":"story"}',
      ),
    ).toEqual({
      entityId: "story-1",
      entityType: "story",
      kind: "notification",
    });
  });
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
    '{"type":"story.workspace_update","storyId":"story-1","changes":{"checklistChanged":false}}',
    '{"type":"story.workspace_update","storyId":"story-1","changes":{"checklistChanged":"true"}}',
  ])("rejects malformed or unsupported payload %s", (payload) => {
    expect(() => parseWorkspaceRealtimeEvent(payload)).toThrow(
      WorkspaceRealtimeContractError,
    );
  });

  it("decodes checklist invalidation alongside ordinary story changes", () => {
    expect(
      parseWorkspaceRealtimeEvent(
        JSON.stringify({
          changes: { checklistChanged: true, title: "Review release" },
          storyId: "story-1",
          type: "story.workspace_update",
        }),
      ),
    ).toEqual({
      changes: { checklistChanged: true, title: "Review release" },
      kind: "story-updated",
      storyId: "story-1",
    });
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

it("decodes scoped chat hints and rejects missing identities", () => {
  const payload = {
    type: "chat.updated",
    eventId: "11111111-1111-4111-8111-111111111111",
    workspaceId: "22222222-2222-4222-8222-222222222222",
    userId: "33333333-3333-4333-8333-333333333333",
    conversationId: "44444444-4444-4444-8444-444444444444",
  };
  expect(parseWorkspaceRealtimeEvent(JSON.stringify(payload))).toEqual({
    kind: "chat-updated",
    workspaceId: payload.workspaceId,
    userId: payload.userId,
    conversationId: payload.conversationId,
  });
  expect(() =>
    parseWorkspaceRealtimeEvent(
      JSON.stringify({ ...payload, userId: undefined }),
    ),
  ).toThrow(WorkspaceRealtimeContractError);
});

it("validates typing scope, identity and expiry without treating it as a notification", () => {
  const payload = {
    type: "chat.typing",
    workspaceId: "11111111-1111-4111-8111-111111111111",
    userId: "22222222-2222-4222-8222-222222222222",
    conversationId: "33333333-3333-4333-8333-333333333333",
    actorId: "44444444-4444-4444-8444-444444444444",
    fullName: "Sarah Chen",
    threadRootId: null,
    isTyping: true,
    expiresAt: "2026-10-16T19:00:06.000Z",
  };
  const { type: _type, ...activity } = payload;
  expect(parseWorkspaceRealtimeEvent(JSON.stringify(payload))).toEqual({
    kind: "chat-typing",
    ...activity,
  });
  for (const invalid of [
    { actorId: null },
    { userId: undefined },
    { threadRootId: "another-thread" },
    { fullName: "" },
    { isTyping: "true" },
    { expiresAt: "never" },
  ]) {
    expect(() =>
      parseWorkspaceRealtimeEvent(JSON.stringify({ ...payload, ...invalid })),
    ).toThrow(WorkspaceRealtimeContractError);
  }
});
