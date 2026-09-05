"use client";

import { useCallback, useEffect, useMemo, useSyncExternalStore } from "react";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import {
  addStoryVisit,
  getRecentStoryHistoryKey,
  parseStoryVisits,
} from "./recent-history";

const HISTORY_CHANGED_EVENT = "recent-story-history-changed";
const memoryHistory = new Map<string, string>();
const getServerSnapshot = () => null;

const readHistory = (key: string | null) => {
  if (!key) return null;
  if (memoryHistory.has(key)) return memoryHistory.get(key) ?? null;
  try {
    return window.localStorage.getItem(key);
  } catch {
    // Keep this session usable when browser storage is unavailable.
    return memoryHistory.get(key) ?? null;
  }
};

const useHistoryKey = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return session?.user.id && workspaceSlug
    ? getRecentStoryHistoryKey(session.user.id, workspaceSlug)
    : null;
};

export const useRecentStoryHistory = () => {
  const key = useHistoryKey();
  const subscribe = useCallback(
    (onChange: () => void) => {
      const onStorage = (event: StorageEvent) => {
        if (event.key === key || event.key === null) onChange();
      };
      const onHistoryChanged = (event: Event) => {
        if ((event as CustomEvent<string>).detail === key) onChange();
      };
      window.addEventListener("storage", onStorage);
      window.addEventListener(HISTORY_CHANGED_EVENT, onHistoryChanged);
      return () => {
        window.removeEventListener("storage", onStorage);
        window.removeEventListener(HISTORY_CHANGED_EVENT, onHistoryChanged);
      };
    },
    [key],
  );
  const getSnapshot = useCallback(() => readHistory(key), [key]);
  const raw = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
  return useMemo(() => parseStoryVisits(raw), [raw]);
};

export const useRecordStoryVisit = (storyId?: string) => {
  const key = useHistoryKey();
  useEffect(() => {
    if (!key || !storyId) return;
    const next = JSON.stringify(
      addStoryVisit(parseStoryVisits(readHistory(key)), {
        storyId,
        visitedAt: new Date().toISOString(),
      }),
    );
    try {
      window.localStorage.setItem(key, next);
      memoryHistory.delete(key);
    } catch {
      // Store only identifiers and timestamps in memory if persistence fails.
      memoryHistory.set(key, next);
    }
    window.dispatchEvent(
      new CustomEvent(HISTORY_CHANGED_EVENT, { detail: key }),
    );
  }, [key, storyId]);
};
