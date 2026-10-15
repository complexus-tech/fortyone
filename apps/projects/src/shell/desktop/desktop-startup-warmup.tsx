"use client";
import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect, useRef } from "react";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { getDesktopBridge } from "./bridge";
import {
  getDesktopOnline,
  DESKTOP_CONNECTIVITY_EVENT,
  probeDesktopConnectivity,
} from "./connectivity";
import { myWorkParams } from "./my-work-warmup-params";
import { warmDesktopMyWork, warmDesktopSettings } from "./startup-queries";
import {
  DESKTOP_CACHE_SAVE_EVENT,
  DESKTOP_CACHE_RETRY_EVENT,
  updateDesktopCacheStatus,
} from "./cache-status";

/** Warm real screen dependencies only inside the existing desktop shell. */
export const DesktopStartupWarmup = ({
  workspaceId,
}: {
  workspaceId: string;
}) => {
  const { data: session } = useSession();
  const { workspaceSlug, withWorkspace } = useWorkspacePath();
  const router = useRouter();
  const queryClient = useQueryClient();
  const warmed = useRef<string | null>(null);
  const myWorkPath = withWorkspace("/my-work");
  const settingsPath = withWorkspace("/settings");
  useEffect(() => {
    if (!getDesktopBridge() || !session?.user.id || !workspaceSlug) return;
    const identity = `${session.user.id}:${workspaceSlug}`;
    let active = true;
    const isActive = () => active;
    let running = false;
    const warm = async () => {
      if (
        !isActive() ||
        running ||
        warmed.current === identity ||
        !getDesktopOnline()
      )
        return;
      running = true;
      router.prefetch(myWorkPath);
      router.prefetch(settingsPath);
      const ctx = { session, workspaceSlug };
      updateDesktopCacheStatus({
        workspaceId,
        phase: "settings",
        downloadError: undefined,
      });
      const failures = await warmDesktopSettings(queryClient, ctx);
      if (!isActive()) return;
      updateDesktopCacheStatus({ workspaceId, phase: "my-work" });
      try {
        await warmDesktopMyWork(queryClient, ctx, myWorkParams(myWorkPath));
      } catch (reason) {
        failures.push(reason);
      }
      if (!isActive()) return;
      running = false;
      if (failures.length) {
        void probeDesktopConnectivity();
        updateDesktopCacheStatus({
          workspaceId,
          phase: "paused",
          downloadError:
            "Some workspace data could not be downloaded. Retry when connected.",
        });
      } else {
        warmed.current = identity;
        updateDesktopCacheStatus({
          workspaceId,
          phase: "ready",
          downloadedAt: Date.now(),
          downloadError: undefined,
        });
        window.dispatchEvent(new Event(DESKTOP_CACHE_SAVE_EVENT));
      }
    };
    const run = () => {
      void warm();
    };
    const retry = () => {
      warmed.current = null;
      run();
    };
    const timer = window.setTimeout(run, 100);
    window.addEventListener(DESKTOP_CONNECTIVITY_EVENT, retry);
    window.addEventListener(DESKTOP_CACHE_RETRY_EVENT, retry);
    return () => {
      active = false;
      window.clearTimeout(timer);
      window.removeEventListener(DESKTOP_CONNECTIVITY_EVENT, retry);
      window.removeEventListener(DESKTOP_CACHE_RETRY_EVENT, retry);
    };
  }, [
    myWorkPath,
    queryClient,
    router,
    session,
    settingsPath,
    workspaceId,
    workspaceSlug,
  ]);
  return null;
};
