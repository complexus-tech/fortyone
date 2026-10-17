import { useEffect, useState } from "react";
import { hasDocumentMayaMention } from "./document-comment-markdown";
import type { DocumentMayaRun } from "./document-comment-maya";
import type { DocumentComment } from "./types";

const PENDING_RUN_GRACE_MS = 15_000;

/** Bridge comment creation to the first run response without an impure render clock. */
export function useDocumentMayaSources(
  comments: DocumentComment[],
  runs: DocumentMayaRun[],
) {
  const [now, setNow] = useState(Date.now);
  const sources = comments.filter(
    (comment) =>
      hasDocumentMayaMention(comment.body) &&
      !runs.some((run) => run.sourceId === comment.id) &&
      (comment.id.startsWith("optimistic-") ||
        now - Date.parse(comment.createdAt) < PENDING_RUN_GRACE_MS),
  );
  const expiresAt = Math.min(
    ...sources
      .filter((comment) => !comment.id.startsWith("optimistic-"))
      .map((comment) => Date.parse(comment.createdAt) + PENDING_RUN_GRACE_MS),
  );
  useEffect(() => {
    if (!Number.isFinite(expiresAt)) return;
    const timer = window.setTimeout(
      () => {
        setNow(Date.now());
      },
      Math.max(0, expiresAt - Date.now()),
    );
    return () => {
      window.clearTimeout(timer);
    };
  }, [expiresAt]);
  return sources;
}
