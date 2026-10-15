"use client";

import { useEffect } from "react";
import { usePathname } from "next/navigation";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { getDesktopBridge } from "./bridge";
import {
  getDesktopOnline,
  DESKTOP_CONNECTIVITY_EVENT,
  probeDesktopConnectivity,
} from "./connectivity";
import { bindDesktopSession } from "./session-binding";
import { updateDesktopCacheStatus } from "./cache-status";
import {
  isDesktopOfflineNavigation,
  navigateWithDesktopOffline,
} from "./offline-navigation";

const STATUS_EVENT = "fortyone:desktop-shell-cache-status";
const RETRY_EVENT = "fortyone:desktop-cache-retry";
const downloaded = new Map<string, number>();
const DOWNLOAD_INTERVAL = 5 * 60_000;
type DownloadedPage = {
  url: string;
  html: string;
  contentSecurityPolicy?: string;
};

function workerRequest(worker: ServiceWorker, message: unknown) {
  return new Promise<DownloadedPage[]>((resolve, reject) => {
    const channel = new MessageChannel();
    const timeout = window.setTimeout(() => {
      channel.port1.close();
      reject(
        new Error(
          "Downloading app pages took too long. Try again when connected.",
        ),
      );
    }, 60_000);
    channel.port1.onmessage = ({
      data,
    }: MessageEvent<{
      ok?: boolean;
      error?: string;
      pages?: DownloadedPage[];
    }>) => {
      window.clearTimeout(timeout);
      channel.port1.close();
      if (data.ok) resolve(data.pages ?? []);
      else reject(new Error(data.error ?? "Could not download the app pages."));
    };
    worker.postMessage(message, [channel.port2]);
  });
}

async function activeWorker() {
  const existing = await navigator.serviceWorker.getRegistration("/");
  if (existing?.active) return existing.active;
  return new Promise<ServiceWorker>((resolve, reject) => {
    const timeout = window.setTimeout(() => {
      reject(
        new Error(
          "Offline page startup could not be prepared. Try again when connected.",
        ),
      );
    }, 15_000);
    void navigator.serviceWorker.ready.then((registration) => {
      window.clearTimeout(timeout);
      if (registration.active) resolve(registration.active);
      else reject(new Error("Offline page startup is unavailable."));
    });
  });
}

/** Download the existing Next interface; private HTML stays behind native account checks. */
export function DesktopOfflineShell({ workspaceId }: { workspaceId: string }) {
  const { data: session } = useSession();
  const { workspaceSlug, withWorkspace } = useWorkspacePath();
  const pathname = usePathname();
  const accountId = session?.user.id;
  const myWorkPath = withWorkspace("/my-work");
  const settingsPath = withWorkspace("/settings");
  const profilePath = withWorkspace("/settings/account");
  const preferencesPath = withWorkspace("/settings/account/preferences");
  const notificationsPath = withWorkspace("/settings/account/notifications");

  useEffect(() => {
    const bridge = getDesktopBridge();
    const restore = bridge?.restoreSession;
    const savePage = bridge?.pageWrite;
    if (!restore || !savePage || !accountId || !workspaceSlug) return;
    let active = true;
    let running = false;
    const isActive = () => active;
    const report = (
      phase: "ready" | "saving" | "unavailable",
      error?: string,
    ) => {
      if (isActive()) {
        updateDesktopCacheStatus({
          workspaceId,
          shellError: phase === "unavailable" ? error : undefined,
        });
        window.dispatchEvent(
          new CustomEvent(STATUS_EVENT, { detail: { phase, error } }),
        );
      }
    };
    const prepare = async (refresh = false) => {
      if (!isActive() || running || !getDesktopOnline()) return;
      if (!("serviceWorker" in navigator)) {
        report(
          "unavailable",
          "This desktop WebView does not support offline page startup. Downloaded data and drafts remain saved.",
        );
        return;
      }
      running = true;
      try {
        report("saving");
        const bound = await bindDesktopSession(
          { accountId, workspaceId, workspaceSlug },
          refresh,
        );
        if (!bound || !isActive()) return;
        const scope = await restore({ workspaceSlug });
        if (
          !isActive() ||
          scope.accountId !== accountId ||
          scope.workspaceId !== workspaceId
        )
          return;
        await navigator.serviceWorker.register("/desktop-worker.js", {
          scope: "/",
          updateViaCache: "none",
        });
        const worker = await activeWorker();
        if (!isActive()) return;
        const urls = [
          ...new Set([
            pathname,
            myWorkPath,
            settingsPath,
            profilePath,
            preferencesPath,
            notificationsPath,
          ]),
        ].filter((url) => {
          if (
            url.startsWith(`${settingsPath}/`) &&
            ![
              profilePath,
              preferencesPath,
              notificationsPath,
              `${settingsPath}/workspace/members`,
            ].includes(url)
          )
            return false;
          const lastSaved =
            downloaded.get(`${accountId}:${workspaceId}:${url}`) ?? 0;
          return refresh || Date.now() - lastSaved > DOWNLOAD_INTERVAL;
        });
        for (let index = 0; index < urls.length; index += 3) {
          if (!isActive()) return;
          const batch = urls.slice(index, index + 3);
          // eslint-disable-next-line no-await-in-loop -- Bound Next page rendering and static downloads to one small desktop-only batch.
          const pages = await workerRequest(worker, {
            type: "warm",
            scope,
            urls: batch,
          });
          for (const page of pages) {
            if (!isActive()) return;
            // Revalidates the current HttpOnly session after the HTTP download.
            // eslint-disable-next-line no-await-in-loop -- Commit each private page only after the preceding native account check finishes.
            await savePage({ accountId, workspaceId, ...page });
          }
          for (const url of batch)
            downloaded.set(`${accountId}:${workspaceId}:${url}`, Date.now());
        }
        report("ready");
      } catch (reason) {
        void probeDesktopConnectivity();
        report(
          "unavailable",
          reason instanceof Error
            ? reason.message
            : "Could not download app pages for offline startup.",
        );
      } finally {
        running = false;
      }
    };
    const retry = () => {
      void prepare(true);
    };
    const clear = () => {
      active = false;
      downloaded.clear();
      const worker =
        "serviceWorker" in navigator
          ? navigator.serviceWorker.controller
          : null;
      if (worker)
        void workerRequest(worker, { type: "clear" }).catch(() => undefined);
    };
    const navigateOffline = (event: MouseEvent) => {
      if (
        getDesktopOnline() ||
        event.defaultPrevented ||
        event.button !== 0 ||
        event.metaKey ||
        event.ctrlKey ||
        event.shiftKey ||
        event.altKey
      )
        return;
      const anchor =
        event.target instanceof Element
          ? event.target.closest("a[href]")
          : null;
      if (
        !(anchor instanceof HTMLAnchorElement) ||
        anchor.target ||
        anchor.download
      )
        return;
      const url = new URL(anchor.href);
      if (!isDesktopOfflineNavigation(url.href)) return;
      event.preventDefault();
      event.stopPropagation();
      void navigateWithDesktopOffline(url.href, () => undefined);
    };
    const timer = window.setTimeout(() => {
      void prepare();
    }, 1_000);
    window.addEventListener(DESKTOP_CONNECTIVITY_EVENT, retry);
    window.addEventListener(RETRY_EVENT, retry);
    window.addEventListener("fortyone:desktop-cache-cleared", clear);
    document.addEventListener("click", navigateOffline, true);
    return () => {
      active = false;
      window.clearTimeout(timer);
      window.removeEventListener(DESKTOP_CONNECTIVITY_EVENT, retry);
      window.removeEventListener(RETRY_EVENT, retry);
      window.removeEventListener("fortyone:desktop-cache-cleared", clear);
      document.removeEventListener("click", navigateOffline, true);
    };
  }, [
    accountId,
    myWorkPath,
    pathname,
    settingsPath,
    profilePath,
    preferencesPath,
    notificationsPath,
    workspaceId,
    workspaceSlug,
  ]);
  return null;
}
