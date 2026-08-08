"use client";

import { useCallback, useEffect } from "react";

const AUTOMATIC_SYNC_STALE_AFTER_MS = 5 * 60 * 1000;
const automaticSyncAttempts = new Map<string, number>();

type SyncCalendar = (input: { connectionId: string; silent?: boolean }) => void;

export const useCalendarAutoSync = ({
  connectionId,
  isSyncPending,
  lastSyncedAt,
  sync,
  syncStatus,
}: {
  connectionId?: string;
  isSyncPending: boolean;
  lastSyncedAt?: string | null;
  sync: SyncCalendar;
  syncStatus?: string;
}) => {
  const markSyncAttempt = useCallback((targetConnectionId: string) => {
    automaticSyncAttempts.set(targetConnectionId, Date.now());
  }, []);

  useEffect(() => {
    if (!connectionId) return;

    const syncIfStale = () => {
      const now = Date.now();
      const lastSync = lastSyncedAt ? Date.parse(lastSyncedAt) : Number.NaN;
      const isFresh =
        syncStatus !== "failed" &&
        Number.isFinite(lastSync) &&
        now - lastSync < AUTOMATIC_SYNC_STALE_AFTER_MS;
      const attemptedAt = automaticSyncAttempts.get(connectionId);
      const attemptedRecently =
        attemptedAt !== undefined &&
        now - attemptedAt < AUTOMATIC_SYNC_STALE_AFTER_MS;

      if (isSyncPending || isFresh || attemptedRecently) return;

      markSyncAttempt(connectionId);
      sync({ connectionId, silent: true });
    };

    syncIfStale();
    const interval = window.setInterval(
      syncIfStale,
      AUTOMATIC_SYNC_STALE_AFTER_MS,
    );

    return () => {
      window.clearInterval(interval);
    };
  }, [
    connectionId,
    isSyncPending,
    lastSyncedAt,
    markSyncAttempt,
    sync,
    syncStatus,
  ]);

  return markSyncAttempt;
};
