"use client";
import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { flushDesktopAttachments } from "./attachments";
import { DESKTOP_CONNECTIVITY_EVENT } from "./connectivity";
import type { DesktopCacheScope } from "./cache-types";
import { retryDesktopDrafts } from "./drafts";
import { getDesktopBridge } from "./bridge";
import { bindDesktopSession, clearDesktopBindings } from "./session-binding";
import {
  createDesktopCache,
  isDesktopCacheKey,
  restoreDesktopCache,
} from "./query-cache";
import {
  DESKTOP_CACHE_SAVE_EVENT,
  DESKTOP_CACHE_RETRY_EVENT,
  updateDesktopCacheStatus,
} from "./cache-status";

const owners = new WeakMap<object, string>();
let activeScope: DesktopCacheScope | null = null;
let stopPersistence: (() => Promise<void>) | null = null;

export const getActiveDesktopCacheScope = () => activeScope;

export async function clearDesktopQueryCache() {
  clearDesktopBindings();
  await stopPersistence?.();
  stopPersistence = null;
  const scope = activeScope;
  activeScope = null;
  try {
    if (scope)
      await getDesktopBridge()?.cacheClear?.({ accountId: scope.accountId });
  } finally {
    window.dispatchEvent(new Event("fortyone:desktop-cache-cleared"));
  }
}

export function DesktopQueryPersistence({
  workspaceId,
  workspaceSlug,
}: {
  workspaceId: string;
  workspaceSlug: string;
}) {
  const { data: session } = useSession();
  const accountId = session?.user.id;
  const client = useQueryClient();
  useEffect(() => {
    const bridge = getDesktopBridge();
    if (
      !accountId ||
      !bridge?.bindSession ||
      !bridge.cacheRead ||
      !bridge.cacheWrite
    )
      return;
    const scope = { accountId, workspaceId, workspaceSlug };
    const requestScope = { accountId, workspaceId, version: 1 as const };
    const previousOwner = owners.get(client);
    if (previousOwner && previousOwner !== accountId) client.clear();
    owners.set(client, accountId);
    activeScope = scope;
    let active = true;
    const isActive = () => active;
    let ready = false;
    let timer: ReturnType<typeof setTimeout> | undefined;
    let writes = Promise.resolve();
    const report = (reason: unknown) => {
      if (isActive())
        updateDesktopCacheStatus({
          workspaceId,
          storageError:
            reason instanceof Error ? reason.message : String(reason),
        });
    };
    const save = () => {
      if (!isActive() || !ready || client.isMutating()) return;
      const payload = createDesktopCache(client, scope);
      writes = writes
        .catch(() => undefined)
        .then(async () => {
          if (!isActive()) return;
          await bridge.cacheWrite!({ ...requestScope, payload });
          if (isActive())
            updateDesktopCacheStatus({
              workspaceId,
              savedAt: payload.savedAt,
              storageError: undefined,
            });
        })
        .catch(report);
    };
    const schedule = () => {
      clearTimeout(timer);
      timer = setTimeout(save, 350);
    };
    const restore = async (refresh = false) => {
      updateDesktopCacheStatus({ workspaceId, phase: "restoring" });
      try {
        await bindDesktopSession(scope, refresh);
        const cached = await bridge.cacheRead!(requestScope);
        if (!isActive()) return;
        restoreDesktopCache(client, scope, cached);
        ready = true;
        schedule();
      } catch (reason) {
        report(reason);
      }
    };
    const unsubscribeQueries = client.getQueryCache().subscribe((event) => {
      if (
        event.type === "updated" &&
        isDesktopCacheKey(event.query.queryKey as readonly unknown[], scope)
      )
        schedule();
    });
    const unsubscribeMutations = client.getMutationCache().subscribe(schedule);
    const retry = () => {
      if (isActive()) void restore(true);
    };
    const beforeClose = (event: Event) => {
      const detail = (
        event as CustomEvent<{ waitUntil: (promise: Promise<void>) => void }>
      ).detail;
      detail.waitUntil(
        Promise.all([
          flushDesktopAttachments(),
          retryDesktopDrafts(scope),
        ]).then(() => undefined),
      );
    };
    window.addEventListener("fortyone:desktop-before-close", beforeClose);
    window.addEventListener(DESKTOP_CACHE_RETRY_EVENT, retry);
    window.addEventListener(DESKTOP_CONNECTIVITY_EVENT, retry);
    window.addEventListener(DESKTOP_CACHE_SAVE_EVENT, schedule);
    const cleanup = () => {
      active = false;
      clearTimeout(timer);
      unsubscribeQueries();
      unsubscribeMutations();
      window.removeEventListener("fortyone:desktop-before-close", beforeClose);
      window.removeEventListener(DESKTOP_CACHE_RETRY_EVENT, retry);
      window.removeEventListener(DESKTOP_CONNECTIVITY_EVENT, retry);
      window.removeEventListener(DESKTOP_CACHE_SAVE_EVENT, schedule);
    };
    const stop = async () => {
      cleanup();
      await writes;
    };
    stopPersistence = stop;
    void restore();
    return () => {
      cleanup();
      if (stopPersistence === stop) stopPersistence = null;
    };
  }, [accountId, client, workspaceId, workspaceSlug]);
  return null;
}
