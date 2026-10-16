/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { act, renderHook } from "@testing-library/react";
import type { Comment } from "@/types/comment";
import type { CommentMayaRun } from "../types/comment-maya";
import { useMayaPendingReplies } from "./use-maya-pending-replies";

const NOW = "2026-10-16T13:00:00Z";
const comment = (id: string, subComments: Comment[] = []): Comment => ({
  id,
  storyId: "story",
  parentId: null,
  userId: "me",
  user: {
    id: "me",
    fullName: "Joseph",
    username: "joseph",
    isActive: true,
    isSystem: false,
    avatarUrl: null,
  },
  comment: "<p>A comment</p>",
  createdAt: NOW,
  updatedAt: NOW,
  subComments,
});
const run: CommentMayaRun = {
  id: "run",
  sourceId: "root",
  requesterId: "me",
  status: "queued",
  summary: "",
  updatedAt: NOW,
  version: 1,
};

beforeEach(() => {
  jest.useFakeTimers();
  jest.setSystemTime(new Date(NOW));
});
afterEach(() => {
  jest.useRealTimers();
});

describe("Maya pending reply presentation", () => {
  it("shows only confirmed runs and maps a nested request to its root without adding comments", () => {
    const comments = [comment("root", [comment("request")])];
    const { result, rerender } = renderHook(
      ({ runs }) => useMayaPendingReplies(comments, runs),
      { initialProps: { runs: [] as CommentMayaRun[] } },
    );
    expect(result.current.size).toBe(0);
    rerender({ runs: [{ ...run, sourceId: "request" }] });
    expect(result.current.get("root")?.[0].status).toBe("thinking");
    expect(result.current.has("request")).toBe(false);
    expect(comments[0].subComments).toHaveLength(1);
  });

  it("keeps the row while a published response is loading and removes it exactly when that reply arrives", () => {
    const { result, rerender } = renderHook(
      ({ comments, runs }) => useMayaPendingReplies(comments, runs),
      { initialProps: { comments: [comment("root")], runs: [run] } },
    );
    const published = {
      ...run,
      status: "running" as const,
      replyCommentId: "reply",
    };
    rerender({ comments: [comment("root")], runs: [published] });
    expect(result.current.get("root")?.[0].status).toBe("replying");
    rerender({
      comments: [comment("root", [comment("reply")])],
      runs: [published],
    });
    expect(result.current.size).toBe(0);
  });

  it("does not restart the waiting clock on retries or new versions and eventually stops animating", () => {
    const { result, rerender } = renderHook(
      ({ runs }) => useMayaPendingReplies([comment("root")], runs),
      { initialProps: { runs: [run] } },
    );
    act(() => {
      jest.advanceTimersByTime(30_000);
    });
    expect(result.current.get("root")?.[0].status).toBe("working");
    rerender({
      runs: [{ ...run, version: 2, updatedAt: new Date().toISOString() }],
    });
    act(() => {
      jest.advanceTimersByTime(150_000);
    });
    expect(result.current.get("root")?.[0].status).toBe("delayed");
    rerender({
      runs: [{ ...run, version: 3, updatedAt: new Date().toISOString() }],
    });
    expect(result.current.get("root")?.[0].status).toBe("delayed");
  });

  it("uses static feedback when polling fails and resumes the same waiting clock on recovery", () => {
    const { result, rerender } = renderHook(
      ({ unavailable }) =>
        useMayaPendingReplies([comment("root")], [run], unavailable),
      { initialProps: { unavailable: false } },
    );
    rerender({ unavailable: true });
    expect(result.current.get("root")?.[0].status).toBe("offline");
    act(() => {
      jest.advanceTimersByTime(180_000);
    });
    rerender({ unavailable: false });
    expect(result.current.get("root")?.[0].status).toBe("delayed");
  });

  it("starts a new waiting cycle after an explicit retry despite the previous failure reply", () => {
    const comments = [comment("root", [comment("old-failure")])];
    const { result, rerender } = renderHook(
      ({ runs }) => useMayaPendingReplies(comments, runs),
      { initialProps: { runs: [run] } },
    );
    act(() => {
      jest.advanceTimersByTime(180_000);
    });
    rerender({
      runs: [{ ...run, status: "failed", replyCommentId: "old-failure" }],
    });
    expect(result.current.size).toBe(0);
    expect(jest.getTimerCount()).toBe(0);
    act(() => {
      jest.advanceTimersByTime(3_600_000);
    });
    rerender({
      runs: [{ ...run, version: 3, updatedAt: new Date().toISOString() }],
    });
    expect(result.current.get("root")?.[0].status).toBe("thinking");
    act(() => {
      jest.advanceTimersByTime(0);
    });
    expect(result.current.get("root")?.[0].status).toBe("thinking");
    act(() => {
      jest.advanceTimersByTime(30_000);
    });
    expect(result.current.get("root")?.[0].status).toBe("working");
  });

  it.each([0, 7_200_000])(
    "observes a fresh clock for a request after a long idle period with %ims server skew",
    (serverClockSkew) => {
      const { result, rerender } = renderHook(
        ({ runs }) => useMayaPendingReplies([comment("root")], runs),
        { initialProps: { runs: [] as CommentMayaRun[] } },
      );
      expect(jest.getTimerCount()).toBe(0);
      act(() => {
        jest.advanceTimersByTime(3_600_000);
      });
      rerender({
        runs: [
          {
            ...run,
            updatedAt: new Date(Date.now() + serverClockSkew).toISOString(),
          },
        ],
      });
      expect(result.current.get("root")?.[0].status).toBe("thinking");
      act(() => {
        jest.advanceTimersByTime(0);
      });
      expect(result.current.get("root")?.[0].status).toBe("thinking");
      act(() => {
        jest.advanceTimersByTime(30_000);
      });
      expect(result.current.get("root")?.[0].status).toBe("working");
      act(() => {
        jest.advanceTimersByTime(150_000);
      });
      expect(result.current.get("root")?.[0].status).toBe("delayed");
    },
  );

  it("recognizes an already delayed request when it becomes visible after an idle period", () => {
    const { result, rerender } = renderHook(
      ({ runs }) => useMayaPendingReplies([comment("root")], runs),
      { initialProps: { runs: [] as CommentMayaRun[] } },
    );
    act(() => {
      jest.advanceTimersByTime(3_600_000);
    });
    rerender({ runs: [run] });
    act(() => {
      jest.advanceTimersByTime(0);
    });
    expect(result.current.get("root")?.[0].status).toBe("delayed");
  });

  it("starts a newly visible request without resetting another request's elapsed time", () => {
    const { result, rerender } = renderHook(
      ({ runs }) => useMayaPendingReplies([comment("root")], runs),
      { initialProps: { runs: [run] } },
    );
    act(() => {
      jest.advanceTimersByTime(150_000);
    });
    rerender({
      runs: [
        run,
        { ...run, id: "new-run", updatedAt: new Date().toISOString() },
      ],
    });
    act(() => {
      jest.advanceTimersByTime(0);
    });
    expect(
      result.current.get("root")?.map((pending) => pending.status),
    ).toEqual(["working", "thinking"]);
    act(() => {
      jest.advanceTimersByTime(30_000);
    });
    expect(
      result.current.get("root")?.map((pending) => pending.status),
    ).toEqual(["delayed", "working"]);
  });

  it("does not show phantom activity for cancelled, completed, or unavailable sources", () => {
    const { result } = renderHook(() =>
      useMayaPendingReplies(
        [comment("root")],
        [
          { ...run, id: "cancelled", status: "cancelled" },
          { ...run, id: "done", status: "succeeded" },
          { ...run, id: "question", status: "needs_input" },
          { ...run, id: "unloaded", sourceId: "another-page" },
        ],
      ),
    );
    expect(result.current.size).toBe(0);
  });

  it("clears its clock when the feed unmounts", () => {
    const { unmount } = renderHook(() =>
      useMayaPendingReplies([comment("root")], [run]),
    );
    act(() => {
      jest.advanceTimersByTime(0);
    });
    expect(jest.getTimerCount()).toBe(1);
    unmount();
    expect(jest.getTimerCount()).toBe(0);
  });

  it("cancels a queued first observation if the request finishes before it runs", () => {
    const { result, rerender } = renderHook(
      ({ runs }) => useMayaPendingReplies([comment("root")], runs),
      { initialProps: { runs: [] as CommentMayaRun[] } },
    );
    rerender({ runs: [run] });
    expect(jest.getTimerCount()).toBe(2);
    rerender({ runs: [{ ...run, status: "succeeded" }] });
    expect(result.current.size).toBe(0);
    expect(jest.getTimerCount()).toBe(0);
  });

  it("caps waiting even when the server clock is ahead", () => {
    const { result } = renderHook(() =>
      useMayaPendingReplies(
        [comment("root")],
        [{ ...run, updatedAt: "2026-10-16T15:00:00Z" }],
      ),
    );
    act(() => {
      jest.advanceTimersByTime(180_000);
    });
    expect(result.current.get("root")?.[0].status).toBe("delayed");
  });
});
