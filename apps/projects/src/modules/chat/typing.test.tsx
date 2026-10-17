import { act, renderHook } from "@testing-library/react";
import { applyChatTypingEvent, clearChatTyping, useChatTyping } from "./typing";
import { createTypingActivity } from "./typing-activity";

const event = (overrides = {}) => ({
  workspaceId: "workspace",
  conversationId: "room",
  actorId: "person",
  fullName: "Sarah",
  threadRootId: null,
  isTyping: true,
  expiresAt: new Date(Date.now() + 6000).toISOString(),
  ...overrides,
});
const flush = async () => {
  await Promise.resolve();
  await Promise.resolve();
  await Promise.resolve();
  await Promise.resolve();
};

describe("typing leases", () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(new Date("2026-10-16T10:00:00Z"));
  });
  afterEach(() => {
    act(() => {
      clearChatTyping("workspace");
      clearChatTyping("other");
    });
    jest.useRealTimers();
  });
  it("isolates threads and workspaces and expires remote hints within six seconds", () => {
    const room = renderHook(() => useChatTyping("workspace", "room"));
    const thread = renderHook(() => useChatTyping("workspace", "room", "root"));
    act(() => {
      applyChatTypingEvent(
        event({ expiresAt: new Date(Date.now() + 60_000).toISOString() }),
      );
      applyChatTypingEvent(
        event({ threadRootId: "root", fullName: "Thread member" }),
      );
      applyChatTypingEvent(
        event({ workspaceId: "other", fullName: "Other workspace" }),
      );
    });
    expect(room.result.current).toEqual([{ id: "person", fullName: "Sarah" }]);
    expect(thread.result.current).toEqual([
      { id: "person", fullName: "Thread member" },
    ]);
    act(() => {
      jest.advanceTimersByTime(6000);
    });
    expect(room.result.current).toEqual([]);
    expect(thread.result.current).toEqual([]);
  });
  it("rejects a delayed older start after an explicit stop", () => {
    const room = renderHook(() => useChatTyping("workspace", "room"));
    const start = event();
    act(() => {
      applyChatTypingEvent(start);
    });
    act(() => {
      jest.advanceTimersByTime(1000);
    });
    act(() => {
      applyChatTypingEvent(event({ isTyping: false }));
    });
    act(() => {
      applyChatTypingEvent(start);
    });
    expect(room.result.current).toEqual([]);
    act(() => {
      applyChatTypingEvent(
        event({ expiresAt: new Date(Date.now() + 6001).toISOString() }),
      );
    });
    expect(room.result.current).toHaveLength(1);
    act(() => {
      clearChatTyping("workspace");
    });
    expect(room.result.current).toEqual([]);
  });
});

describe("typing activity", () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(0);
  });
  afterEach(() => jest.useRealTimers());
  it("throttles refreshes and stops after idle", async () => {
    const send = jest.fn().mockResolvedValue(undefined);
    const activity = createTypingActivity(send);
    activity.changed();
    await flush();
    jest.advanceTimersByTime(1000);
    activity.changed();
    expect(send.mock.calls).toEqual([[true]]);
    jest.advanceTimersByTime(2000);
    activity.changed();
    await flush();
    expect(send.mock.calls).toEqual([[true], [true]]);
    jest.advanceTimersByTime(4000);
    await flush();
    expect(send.mock.calls).toEqual([[true], [true], [false]]);
    activity.dispose();
  });
  it("coalesces slow requests and discards stale starts on disposal", async () => {
    let finish: () => void = () => {};
    const send = jest.fn(
      () =>
        new Promise<void>((resolve) => {
          finish = resolve;
        }),
    );
    const activity = createTypingActivity(send);
    activity.changed();
    for (let i = 0; i < 10; i++) {
      jest.advanceTimersByTime(3000);
      activity.changed();
    }
    expect(send.mock.calls).toEqual([[true]]);
    activity.dispose();
    activity.changed();
    finish();
    await flush();
    expect(send.mock.calls).toEqual([[true], [false]]);
    finish();
    await flush();
    expect(send).toHaveBeenCalledTimes(2);
  });
});
