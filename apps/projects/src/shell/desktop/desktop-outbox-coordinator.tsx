"use client";

import { useEffect, useSyncExternalStore } from "react";
import { useQueryClient } from "@tanstack/react-query";
import type { DesktopCacheScope } from "./cache-types";
import type {
  DesktopOutboxSnapshot,
  DesktopResolveOperation,
} from "./outbox-types";
import { DESKTOP_OUTBOX_CHANGED } from "./outbox-types";
import { getDesktopBridge } from "./bridge";
import {
  getDesktopOnline,
  DESKTOP_CONNECTIVITY_EVENT,
  probeDesktopConnectivity,
} from "./connectivity";
import { bindDesktopSession } from "./session-binding";
import { useDesktop } from "./desktop-provider";
import { useDesktopWorkspaceScope } from "./workspace-scope";
import { reconcileDesktopOutbox } from "./outbox-projection";
import { discardDesktopProjection } from "./outbox-discard";

type OutboxState = {
  scope: DesktopCacheScope;
  snapshot: DesktopOutboxSnapshot;
  busy: boolean;
  online: boolean;
  error?: string;
};
type Resolution = Omit<DesktopResolveOperation, "accountId" | "workspaceId">;
let state: OutboxState | null = null;
const listeners = new Set<() => void>();
let controller: {
  scope: DesktopCacheScope;
  refresh: () => Promise<void>;
  resolve: (request: Resolution) => Promise<void>;
} | null = null;
const matches = (
  a: DesktopCacheScope | null | undefined,
  b: DesktopCacheScope | null | undefined,
) =>
  Boolean(
    a && b && a.accountId === b.accountId && a.workspaceId === b.workspaceId,
  );
const subscribe = (listener: () => void) => {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
};
const publish = (next: OutboxState | null) => {
  state = next;
  for (const listener of listeners) listener();
};
export const useDesktopOutbox = () => {
  const scope = useDesktopWorkspaceScope();
  return useSyncExternalStore(
    subscribe,
    () => (matches(state?.scope, scope) ? state : null),
    () => null,
  );
};
export const refreshDesktopOutbox = (scope: DesktopCacheScope) =>
  matches(controller?.scope, scope) ? controller!.refresh() : Promise.resolve();
export const resolveDesktopOperation = (
  scope: DesktopCacheScope,
  request: Resolution,
) => {
  if (!matches(controller?.scope, scope))
    return Promise.reject(
      new Error("Return to this workspace before resolving its changes."),
    );
  return controller!.resolve(request);
};

export const DesktopOutboxCoordinator = () => {
  const scope = useDesktopWorkspaceScope();
  const { isDesktop } = useDesktop();
  const client = useQueryClient();
  useEffect(() => {
    const bridge = getDesktopBridge();
    if (
      !isDesktop ||
      !scope ||
      !bridge?.readOutbox ||
      !bridge.syncOutbox ||
      !bridge.resolveOperation
    )
      return;
    const requestScope = {
      accountId: scope.accountId,
      workspaceId: scope.workspaceId,
    };
    let active = true;
    const isActive = () => active;
    let flight: Promise<void> | null = null;
    let rerun = false;
    let projecting = false;
    let projectionTimer: ReturnType<typeof setTimeout> | undefined;
    let snapshot: DesktopOutboxSnapshot = { operations: [], completed: [] };
    const acknowledged = new Set<string>();
    const update = (patch: Partial<OutboxState>) => {
      if (isActive())
        publish({
          scope,
          snapshot,
          busy: false,
          online: getDesktopOnline(),
          ...patch,
        });
    };
    const project = () => {
      if (!isActive() || projecting) return;
      projecting = true;
      try {
        reconcileDesktopOutbox(client, scope, snapshot);
      } finally {
        projecting = false;
      }
    };
    const accept = (next: DesktopOutboxSnapshot) => {
      if (!isActive()) return;
      snapshot = next;
      project();
      const fresh = next.completed.filter(
        (ack) => !acknowledged.has(ack.operationId),
      );
      for (const ack of fresh) acknowledged.add(ack.operationId);
      if (fresh.length)
        void client.invalidateQueries({
          queryKey: ["stories", scope.workspaceSlug],
        });
      update({ busy: true });
    };
    const report = (error: unknown) => {
      void probeDesktopConnectivity();
      update({ error: error instanceof Error ? error.message : String(error) });
    };
    const refresh = (): Promise<void> => {
      if (!isActive()) return Promise.resolve();
      if (flight) {
        rerun = true;
        return flight;
      }
      update({ busy: true });
      const task = (async () => {
        if (!(await bindDesktopSession(scope)) || !isActive()) return;
        accept(await bridge.readOutbox!(requestScope));
        if (
          isActive() &&
          getDesktopOnline() &&
          snapshot.operations.some(
            (operation) => operation.status === "pending",
          )
        ) {
          accept(await bridge.syncOutbox!(requestScope));
        }
      })()
        .catch(report)
        .finally(() => {
          if (isActive()) update({ error: state?.error });
          flight = null;
          if (rerun && isActive()) {
            rerun = false;
            void refresh();
          }
        });
      flight = task;
      return task;
    };
    const resolve = async (request: Resolution) => {
      if (!isActive() || flight)
        throw new Error("Wait for the current synchronization to finish.");
      update({ busy: true });
      const operation = snapshot.operations.find(
        (item) => item.id === request.operationId,
      );
      if (
        request.resolution === "discard" &&
        operation?.kind === "story.update" &&
        !operation.serverValue
      ) {
        update({});
        throw new Error(
          "Reconnect to load the server version before discarding this change.",
        );
      }
      const task = (async () => {
        if (!(await bindDesktopSession(scope)) || !isActive())
          throw new Error("The current workspace changed.");
        const next = await bridge.resolveOperation!({
          ...requestScope,
          ...request,
        });
        if (isActive() && request.resolution === "discard" && operation) {
          projecting = true;
          try {
            discardDesktopProjection(client, scope, operation);
          } finally {
            projecting = false;
          }
        }
        accept(next);
        if (isActive())
          void client.invalidateQueries({
            queryKey: ["stories", scope.workspaceSlug],
          });
      })();
      flight = task;
      try {
        await task;
      } catch (error) {
        report(error);
        throw error;
      } finally {
        flight = null;
        if (isActive()) update({ error: state?.error });
      }
      if (isActive() && (request.resolution === "retry" || rerun)) {
        rerun = false;
        void refresh();
      }
    };
    const changed = (event: Event) => {
      if (!(event instanceof CustomEvent)) return;
      const detail = event.detail as
        | { accountId?: string; workspaceId?: string }
        | undefined;
      if (
        detail?.accountId === scope.accountId &&
        detail.workspaceId === scope.workspaceId
      )
        void refresh();
    };
    const reconnect = () => {
      update({});
      if (getDesktopOnline()) void refresh();
    };
    const foreground = () => {
      if (document.visibilityState === "visible" && getDesktopOnline())
        void refresh();
    };
    const unsubscribe = client.getQueryCache().subscribe((event) => {
      if (
        projecting ||
        event.type !== "updated" ||
        event.query.queryKey[0] !== "stories" ||
        event.query.queryKey[1] !== scope.workspaceSlug
      )
        return;
      clearTimeout(projectionTimer);
      projectionTimer = setTimeout(project, 0);
    });
    const current = { scope, refresh, resolve };
    controller = current;
    update({});
    void refresh();
    const interval = setInterval(() => {
      if (
        getDesktopOnline() &&
        snapshot.operations.some((operation) => operation.status === "pending")
      )
        void refresh();
    }, 30_000);
    window.addEventListener(DESKTOP_OUTBOX_CHANGED, changed);
    window.addEventListener("fortyone:desktop-outbox-change", changed);
    window.addEventListener("online", reconnect);
    window.addEventListener(DESKTOP_CONNECTIVITY_EVENT, reconnect);
    window.addEventListener("offline", reconnect);
    window.addEventListener("focus", foreground);
    document.addEventListener("visibilitychange", foreground);
    return () => {
      active = false;
      clearInterval(interval);
      clearTimeout(projectionTimer);
      unsubscribe();
      window.removeEventListener(DESKTOP_OUTBOX_CHANGED, changed);
      window.removeEventListener("fortyone:desktop-outbox-change", changed);
      window.removeEventListener("online", reconnect);
      window.removeEventListener(DESKTOP_CONNECTIVITY_EVENT, reconnect);
      window.removeEventListener("offline", reconnect);
      window.removeEventListener("focus", foreground);
      document.removeEventListener("visibilitychange", foreground);
      if (controller === current) {
        controller = null;
        publish(null);
      }
    };
  }, [client, isDesktop, scope]);
  return null;
};
