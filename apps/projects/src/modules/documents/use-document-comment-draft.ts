"use client";

import { useCallback, useEffect, useSyncExternalStore } from "react";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useDesktopWorkspaceScope } from "@/shell/desktop/workspace-scope";

type Draft = {
  version: 1;
  body: string;
  creationId?: string;
  replyToCommentId?: string;
};
type DraftSnapshot = {
  body: string;
  ready: boolean;
  error: string | null;
  creationId?: string;
  replyToCommentId?: string;
};
const EMPTY: DraftSnapshot = { body: "", ready: false, error: null };
const snapshots = new Map<string, DraftSnapshot>();
const listeners = new Map<string, Set<() => void>>();

const parseDraft = (value: string | null): Draft | null => {
  if (!value) return null;
  try {
    const draft: unknown = JSON.parse(value);
    if (
      draft &&
      typeof draft === "object" &&
      "version" in draft &&
      draft.version === 1 &&
      "body" in draft &&
      typeof draft.body === "string"
    ) {
      return {
        version: 1,
        body: draft.body,
        replyToCommentId:
          "replyToCommentId" in draft &&
          typeof draft.replyToCommentId === "string"
            ? draft.replyToCommentId
            : undefined,
        creationId:
          "creationId" in draft && typeof draft.creationId === "string"
            ? draft.creationId
            : undefined,
      };
    }
  } catch {
    // An obsolete or corrupt draft must not prevent a new comment.
  }
  return null;
};

function publish(key: string, next: DraftSnapshot) {
  const previous = snapshots.get(key);
  if (
    previous?.body === next.body &&
    previous.ready === next.ready &&
    previous.error === next.error &&
    previous.creationId === next.creationId &&
    previous.replyToCommentId === next.replyToCommentId
  )
    return;
  snapshots.set(key, next);
  listeners.get(key)?.forEach((listener) => {
    listener();
  });
}

function recover(key: string) {
  try {
    const saved = parseDraft(localStorage.getItem(key));
    publish(key, {
      body: saved?.body ?? "",
      creationId: saved?.creationId,
      replyToCommentId: saved?.replyToCommentId,
      ready: true,
      error: null,
    });
  } catch {
    publish(key, {
      body: snapshots.get(key)?.body ?? "",
      creationId: snapshots.get(key)?.creationId,
      replyToCommentId: snapshots.get(key)?.replyToCommentId,
      ready: true,
      error: "Your saved draft could not be recovered on this device.",
    });
  }
}

export function useDocumentCommentScope() {
  const desktop = useDesktopWorkspaceScope();
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return {
    accountId: desktop?.accountId ?? session?.user.id,
    workspaceId: desktop?.workspaceId ?? workspaceSlug,
  };
}

/** A shared draft store keeps remounted composers in sync with in-flight sends. */
export function useDocumentCommentDraft(documentId: string, target: string) {
  const scope = useDocumentCommentScope();
  const key =
    scope.accountId && scope.workspaceId
      ? `fortyone:document-comment-draft:v1:${JSON.stringify([scope.accountId, scope.workspaceId, documentId, target])}`
      : null;
  const subscribe = useCallback(
    (listener: () => void) => {
      if (!key) return () => undefined;
      const subscribers = listeners.get(key) ?? new Set<() => void>();
      subscribers.add(listener);
      listeners.set(key, subscribers);
      return () => {
        subscribers.delete(listener);
        if (!subscribers.size) listeners.delete(key);
      };
    },
    [key],
  );
  const getSnapshot = useCallback(
    () => (key ? snapshots.get(key) ?? EMPTY : EMPTY),
    [key],
  );
  const state = useSyncExternalStore(subscribe, getSnapshot, () => EMPTY);

  useEffect(() => {
    if (!key) return;
    recover(key);
    const onStorage = (event: StorageEvent) => {
      if (event.key === key || event.key === null) recover(key);
    };
    window.addEventListener("storage", onStorage);
    return () => {
      window.removeEventListener("storage", onStorage);
    };
  }, [key]);

  const save = useCallback(
    (body: string, replyToCommentId?: string) => {
      if (!key) return;
      const current = snapshots.get(key);
      const creationId =
        current?.body === body &&
        current.replyToCommentId === replyToCommentId &&
        current.creationId
          ? current.creationId
          : crypto.randomUUID();
      let error: string | null = null;
      try {
        localStorage.setItem(
          key,
          JSON.stringify({ version: 1, body, creationId, replyToCommentId }),
        );
      } catch {
        error =
          "Your draft could not be saved on this device. Keep this comment open until it sends.";
      }
      publish(key, { body, ready: true, error, creationId, replyToCommentId });
      return creationId;
    },
    [key],
  );

  const clear = useCallback(
    (submittedBody: string, creationId?: string) => {
      if (!key) return;
      const current = snapshots.get(key);
      // Protect text typed after submission, including in another composer/tab.
      if (
        current &&
        (current.body !== submittedBody ||
          (creationId && current.creationId !== creationId))
      )
        return;
      try {
        const persisted = parseDraft(localStorage.getItem(key));
        if (
          persisted &&
          (persisted.body !== submittedBody ||
            (creationId && persisted.creationId !== creationId))
        ) {
          recover(key);
          return;
        }
        localStorage.removeItem(key);
        publish(key, { body: "", ready: true, error: null });
      } catch {
        publish(key, {
          body: current?.body ?? submittedBody,
          creationId: current?.creationId,
          replyToCommentId: current?.replyToCommentId,
          ready: true,
          error:
            "Your comment was sent, but the saved draft could not be removed.",
        });
      }
    },
    [key],
  );

  return { ...state, key, save, clear };
}
