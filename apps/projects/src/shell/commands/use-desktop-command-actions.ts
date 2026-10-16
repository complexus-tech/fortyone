"use client";

import { useEffect, useRef, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { toast } from "sonner";
import { logOut } from "@/components/shared/sidebar/actions";
import { useOptionalSidebar } from "@/components/shared/sidebar/sidebar-context";
import { clearAllStorage } from "@/components/shared/sidebar/utils";
import { useWalkthrough } from "@/components/walkthrough/walkthrough-provider";
import { useFeatures } from "@/hooks/features";
import { useLocalStorage } from "@/hooks/local-storage";
import { useUserRole } from "@/hooks/role";
import { useAnalytics } from "@/hooks/tracking";
import { useTerminology } from "@/hooks/use-terminology-display";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import { getDesktopBridge } from "@/shell/desktop/bridge";
import { useDesktop, useDesktopAction } from "@/shell/desktop/desktop-provider";
import { flushDesktopEditors } from "@/shell/desktop/flush-editors";
import { navigateWithDesktopOffline } from "@/shell/desktop/offline-navigation";
import { useDesktopMenu } from "@/shell/desktop/use-desktop-menu";

type DesktopCommandActions = {
  createStory: () => void;
  createObjective: () => void;
  showKeyboardShortcuts: () => void;
};

async function logOutFromDesktop(resetAnalytics: () => void) {
  const body = document.body;
  const previouslyInert = body.inert;
  body.inert = true;
  try {
    await flushDesktopEditors();
    await logOut();
    resetAnalytics();
    clearAllStorage();
    const mainDomain =
      process.env.NEXT_PUBLIC_DOMAIN === "fortyone.app"
        ? "https://fortyone.app"
        : "/";
    window.location.href = `${mainDomain}?signedOut=true`;
  } catch (error) {
    toast.error(
      error instanceof Error
        ? error.message
        : "Save your local drafts before signing out.",
    );
  } finally {
    body.inert = previouslyInert;
  }
}

export const useDesktopCommandActions = ({
  createStory,
  createObjective,
  showKeyboardShortcuts,
}: DesktopCommandActions) => {
  const { isDesktop } = useDesktop();
  const { workspace } = useCurrentWorkspace();
  const { userRole } = useUserRole();
  const features = useFeatures();
  const { getTermDisplay } = useTerminology();
  const sidebar = useOptionalSidebar();
  const { withWorkspace } = useWorkspacePath();
  const router = useRouter();
  const pathname = usePathname();
  const [, setPathBeforeSettings] = useLocalStorage("pathBeforeSettings", "");
  const { analytics } = useAnalytics();
  const logoutPending = useRef(false);
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const { steps, startWalkthrough } = useWalkthrough();
  const ready = Boolean(workspace?.id && userRole) && !isLoggingOut;
  const canCreateStory =
    ready && (userRole === "admin" || userRole === "member");
  const canCreateObjective =
    canCreateStory && !features.isPending && features.objectiveEnabled;
  const tourIntentKey = `fortyone:desktop-product-tour:${workspace?.id ?? ""}`;

  useDesktopMenu({
    ready,
    canCreateStory,
    canCreateObjective,
    canToggleSidebar: ready && Boolean(sidebar),
    sidebarCollapsed: sidebar?.isCollapsed ?? false,
    storyLabel: getTermDisplay("storyTerm"),
    objectiveLabel: getTermDisplay("objectiveTerm"),
  });

  useDesktopAction("create-story", () => {
    if (canCreateStory && !logoutPending.current) createStory();
  });
  useDesktopAction("create-objective", () => {
    if (canCreateObjective && !logoutPending.current) createObjective();
  });
  useDesktopAction("keyboard-shortcuts", () => {
    if (ready && !logoutPending.current) showKeyboardShortcuts();
  });
  useDesktopAction("toggle-sidebar", () => {
    if (ready && !logoutPending.current) sidebar?.toggleSidebar();
  });
  useDesktopAction("settings", () => {
    if (!ready || logoutPending.current) return;
    const settingsRoot = withWorkspace("/settings");
    const destination = withWorkspace("/settings/account/preferences");
    if (pathname === destination) return;
    if (pathname !== settingsRoot && !pathname.startsWith(`${settingsRoot}/`)) {
      setPathBeforeSettings(
        `${window.location.pathname}${window.location.search}`,
      );
    }
    void navigateWithDesktopOffline(destination, () => {
      router.push(destination);
    });
  });
  useDesktopAction("logout", () => {
    if (!ready || logoutPending.current) return;
    logoutPending.current = true;
    setIsLoggingOut(true);
    void logOutFromDesktop(() => {
      analytics.logout(true);
    }).finally(() => {
      logoutPending.current = false;
      setIsLoggingOut(false);
    });
  });

  useDesktopAction("product-tour", () => {
    if (!ready || logoutPending.current) return;
    if (sidebar && steps.length > 0) {
      startWalkthrough();
      return;
    }
    // The walkthrough overlay lives in the app layout. Keep the intent across
    // a saved-page navigation, then start once that layout has its tour steps.
    sessionStorage.setItem(tourIntentKey, "pending");
    const destination = withWorkspace("/my-work");
    void navigateWithDesktopOffline(destination, () => {
      router.push(destination);
    });
  });
  useEffect(() => {
    if (!isDesktop || !ready || !sidebar || steps.length === 0) return;
    if (sessionStorage.getItem(tourIntentKey) !== "pending") return;
    sessionStorage.removeItem(tourIntentKey);
    startWalkthrough();
  }, [
    isDesktop,
    ready,
    sidebar,
    startWalkthrough,
    steps.length,
    tourIntentKey,
  ]);

  return {
    ready,
    ownsNativeShortcuts: isDesktop && Boolean(getDesktopBridge()?.syncMenu),
  };
};
