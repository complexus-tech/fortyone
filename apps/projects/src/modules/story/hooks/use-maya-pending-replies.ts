import { useEffect, useMemo, useState } from "react";
import type { Comment } from "@/types/comment";
import { flattenComments } from "../lib/comment-navigation";
import type {
  CommentMayaPendingStatus,
  CommentMayaRun,
} from "../types/comment-maya";
import { isMayaRunActive } from "../types/comment-maya";

export type MayaPendingReply = {
  run: CommentMayaRun;
  status: CommentMayaPendingStatus;
};

const WORKING_AFTER_MS = 30_000;
const DELAYED_AFTER_MS = 180_000;
const EMPTY_RUNS: CommentMayaRun[] = [];

type PendingRunStart = { requestedAt: number; startedAt: number | null };

const capRunStart = (requestedAt: number, observedAt: number) =>
  Number.isFinite(requestedAt) ? Math.min(requestedAt, observedAt) : observedAt;

const usePendingRunClock = (runs: CommentMayaRun[]) => {
  const key = runs
    .map((run) => run.id)
    .sort()
    .join(":");
  const [clock, setClock] = useState(() => {
    const now = Date.now();
    return {
      key,
      now,
      starts: new Map<string, PendingRunStart>(
        runs.map((run) => {
          const requestedAt = Date.parse(run.updatedAt);
          return [
            run.id,
            { requestedAt, startedAt: capRunStart(requestedAt, now) },
          ];
        }),
      ),
    };
  });
  if (clock.key !== key) {
    // Retain continuous requests across delivery retries. New requests wait for
    // a fresh clock sample, rather than using time from before an idle period.
    setClock({
      ...clock,
      key,
      starts: new Map(
        runs.map((run) => [
          run.id,
          clock.starts.get(run.id) ?? {
            requestedAt: Date.parse(run.updatedAt),
            startedAt: null,
          },
        ]),
      ),
    });
  }

  const hasPending = runs.length > 0;
  useEffect(() => {
    if (!hasPending) return;
    const observeClock = () => {
      const now = Date.now();
      setClock((current) => ({
        ...current,
        now,
        starts: new Map(
          Array.from(current.starts, ([id, start]) => [
            id,
            start.startedAt === null
              ? { ...start, startedAt: capRunStart(start.requestedAt, now) }
              : start,
          ]),
        ),
      }));
    };
    const initialObservation = window.setTimeout(observeClock, 0);
    const timer = window.setInterval(observeClock, 5_000);
    return () => {
      window.clearTimeout(initialObservation);
      window.clearInterval(timer);
    };
  }, [hasPending, key]);

  return clock;
};

/** Presentation only: pending replies never enter the comment cache or counts. */
export const useMayaPendingReplies = (
  comments: Comment[],
  runs: CommentMayaRun[] = EMPTY_RUNS,
  unavailable = false,
) => {
  const candidates = useMemo(() => {
    const threadByCommentId = new Map<string, string>();
    for (const thread of comments) {
      for (const comment of flattenComments([thread])) {
        threadByCommentId.set(comment.id, thread.id);
      }
    }
    return runs.flatMap((run) => {
      const threadId = threadByCommentId.get(run.sourceId);
      if (
        !threadId ||
        run.status === "cancelled" ||
        (run.replyCommentId && threadByCommentId.has(run.replyCommentId)) ||
        (!isMayaRunActive(run) &&
          !run.replyCommentId &&
          run.status !== "failed")
      )
        return [];
      return [{ run, threadId }];
    });
  }, [comments, runs]);
  const timedRuns = candidates
    .map(({ run }) => run)
    .filter((run) => run.status !== "failed" || run.replyCommentId);
  const clock = usePendingRunClock(timedRuns);

  const pendingByThreadId = new Map<string, MayaPendingReply[]>();
  for (const { run, threadId } of candidates) {
    const startedAt = clock.starts.get(run.id)?.startedAt;
    const elapsed = clock.now - (startedAt ?? clock.now);
    let status: CommentMayaPendingStatus;
    if (run.status === "failed" && !run.replyCommentId) status = "failed";
    else if (unavailable) status = "offline";
    else if (elapsed >= DELAYED_AFTER_MS) status = "delayed";
    else if (run.replyCommentId) status = "replying";
    else if (elapsed >= WORKING_AFTER_MS) status = "working";
    else status = "thinking";
    const pending = pendingByThreadId.get(threadId) ?? [];
    pending.push({ run, status });
    pendingByThreadId.set(threadId, pending);
  }
  return pendingByThreadId;
};
