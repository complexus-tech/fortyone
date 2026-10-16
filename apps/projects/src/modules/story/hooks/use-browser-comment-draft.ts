"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import { useDesktopWorkspaceScope } from "@/shell/desktop/workspace-scope";

type CommentDraft = { version: 1; html: string };

type DraftState = {
  key: string | null;
  recovered: CommentDraft | null;
  error: string | null;
};

function readDraft(key: string): CommentDraft | null {
  const stored = window.localStorage.getItem(key);
  if (!stored) return null;

  let payload: unknown;
  try {
    payload = JSON.parse(stored);
  } catch {
    // An invalid or older draft must not prevent composing a comment.
    return null;
  }

  if (
    payload &&
    typeof payload === "object" &&
    "version" in payload &&
    payload.version === 1 &&
    "html" in payload &&
    typeof payload.html === "string"
  ) {
    return { version: 1, html: payload.html };
  }

  return null;
}

export function useBrowserCommentDraft(key: string, enabled = true) {
  // The authenticated workspace layout provides this scope in both shells.
  const scope = useDesktopWorkspaceScope();
  const storageKey =
    enabled && scope
      ? `fortyone:comment-draft:v1:${JSON.stringify([
          scope.accountId,
          scope.workspaceId,
          key,
        ])}`
      : null;
  const [state, setState] = useState<DraftState>({
    key: null,
    recovered: null,
    error: null,
  });
  const reportedError = useRef<string | null>(null);
  const current = state.key === storageKey;
  const error = current ? state.error : null;

  useEffect(() => {
    if (error && reportedError.current !== error) toast.error(error);
    reportedError.current = error;
  }, [error]);

  useEffect(() => {
    if (!storageKey) return;

    try {
      setState({
        key: storageKey,
        recovered: readDraft(storageKey),
        error: null,
      });
    } catch {
      setState({
        key: storageKey,
        recovered: null,
        error: "Your comment draft could not be recovered from this browser.",
      });
    }
  }, [storageKey]);

  const save = useCallback(
    async (payload: CommentDraft) => {
      if (!storageKey) return;
      try {
        // Write immediately so switching tabs or reloading cannot lose a debounce.
        window.localStorage.setItem(storageKey, JSON.stringify(payload));
        setState((currentState) =>
          currentState.key === storageKey &&
          !currentState.recovered &&
          !currentState.error
            ? currentState
            : { key: storageKey, recovered: null, error: null },
        );
      } catch (reason) {
        setState((current) => ({
          ...current,
          key: storageKey,
          recovered: null,
          error: "Your comment draft could not be saved in this browser.",
        }));
        throw reason;
      }
    },
    [storageKey],
  );

  const discard = useCallback(
    async (expectedHtml?: string) => {
      if (!storageKey) return;

      // A request can finish after another composer has saved a newer draft.
      if (
        expectedHtml !== undefined &&
        readDraft(storageKey)?.html !== expectedHtml
      ) {
        return;
      }

      window.localStorage.removeItem(storageKey);
      setState((current) =>
        current.key === storageKey
          ? { ...current, recovered: null, error: null }
          : current,
      );
    },
    [storageKey],
  );

  const acceptRecovered = useCallback(() => {
    setState((current) =>
      current.key === storageKey ? { ...current, recovered: null } : current,
    );
  }, [storageKey]);

  return {
    ready: !storageKey || current,
    recovered: current ? state.recovered : null,
    error,
    save,
    discard,
    acceptRecovered,
  };
}
