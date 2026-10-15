import { useCallback, useEffect, useRef, useState } from "react";
import { flushDesktopAttachments } from "@/shell/desktop/attachments";
import { useDesktopWorkspaceScope } from "@/shell/desktop/workspace-scope";
import { retryDesktopDrafts } from "@/shell/desktop/drafts";
import type { DesktopWorkspaceDestination } from "@/shell/desktop/bridge";
import { getDesktopBridge } from "@/shell/desktop/bridge";
import { buildWorkspaceUrl } from "@/utils/workspace-url";
import { changeWorkspace } from "./actions";

const PREFERENCE_TIMEOUT_MS = 1_500;
const NATIVE_HANDOFF_TIMEOUT_MS = 1_000;
const NAVIGATION_TIMEOUT_MS = 20_000;

type WorkspaceTarget = { id: string; name: string; slug: string };
type SwitchState = {
  destination: DesktopWorkspaceDestination;
  status: "pending" | "failed";
  error?: string;
};
type SwitchAttempt = {
  destination: DesktopWorkspaceDestination;
  controller: AbortController;
  timer: ReturnType<typeof setTimeout>;
};

const reportBridgeError = (error: unknown) => {
  // eslint-disable-next-line no-console -- Surface optional native failures without blocking hosted navigation.
  console.error("Could not synchronize the desktop workspace switch", error);
};

const waitWithin = async (
  operation: Promise<unknown>,
  milliseconds: number,
  onTimeout?: () => void,
) => {
  let timer: ReturnType<typeof setTimeout> | undefined;
  try {
    await Promise.race([
      operation,
      new Promise<void>((resolve) => {
        timer = setTimeout(() => {
          onTimeout?.();
          resolve();
        }, milliseconds);
      }),
    ]);
  } finally {
    clearTimeout(timer);
  }
};

export const useWorkspaceSwitch = (currentWorkspaceId?: string) => {
  const scope = useDesktopWorkspaceScope();
  const [state, setState] = useState<SwitchState | null>(null);
  const attemptRef = useRef<SwitchAttempt | null>(null);

  const clearAttempt = useCallback(() => {
    const attempt = attemptRef.current;
    attemptRef.current = null;
    if (!attempt) return;
    clearTimeout(attempt.timer);
    attempt.controller.abort();
  }, []);

  useEffect(() => {
    if (!currentWorkspaceId) return;
    const bridge = getDesktopBridge();
    void bridge
      ?.workspaceReady?.({ id: currentWorkspaceId })
      .catch(reportBridgeError);
  }, [currentWorkspaceId]);

  useEffect(() => {
    // A history restoration can revive the source document after navigation.
    const resetRestoredPage = (event: PageTransitionEvent) => {
      if (!event.persisted) return;
      clearAttempt();
      setState(null);
    };
    window.addEventListener("pageshow", resetRestoredPage);
    return () => {
      clearAttempt();
      window.removeEventListener("pageshow", resetRestoredPage);
    };
  }, [clearAttempt]);

  const begin = async (destination: DesktopWorkspaceDestination) => {
    if (attemptRef.current) return;
    const controller = new AbortController();
    const attempt: SwitchAttempt = {
      destination,
      controller,
      timer: setTimeout(() => {
        if (attemptRef.current !== attempt) return;
        controller.abort();
        setState({ destination, status: "failed" });
      }, NAVIGATION_TIMEOUT_MS),
    };
    attemptRef.current = attempt;
    setState({ destination, status: "pending" });

    try {
      if (scope) {
        await flushDesktopAttachments();
        await retryDesktopDrafts(scope);
      }
    } catch (error) {
      clearTimeout(attempt.timer);
      setState({
        destination,
        status: "failed",
        error:
          error instanceof Error
            ? error.message
            : "Save your drafts before switching workspaces.",
      });
      return;
    }
    if (attemptRef.current !== attempt) return;
    const bridge = getDesktopBridge();
    const nativeHandoff = bridge?.beginWorkspaceSwitch
      ? Promise.resolve().then(() => bridge.beginWorkspaceSwitch?.(destination))
      : null;
    await Promise.all([
      nativeHandoff
        ? waitWithin(
            nativeHandoff.catch(reportBridgeError),
            NATIVE_HANDOFF_TIMEOUT_MS,
          )
        : Promise.resolve(),
      // This endpoint records a preference. Destination routes still enforce access.
      waitWithin(
        changeWorkspace(destination.id, controller.signal).catch(() => {
          // Last-used preference failures do not prevent authorized navigation.
        }),
        PREFERENCE_TIMEOUT_MS,
        () => {
          controller.abort();
        },
      ),
    ]);
    if (attemptRef.current !== attempt) return;

    try {
      window.location.assign(destination.url);
    } catch {
      clearTimeout(attempt.timer);
      setState({ destination, status: "failed" });
    }
  };

  const switchWorkspace = (workspace: WorkspaceTarget) => {
    if (workspace.id === currentWorkspaceId || attemptRef.current) return;
    void begin({
      id: workspace.id,
      name: workspace.name,
      url: buildWorkspaceUrl(workspace.slug),
    });
  };

  const retry = () => {
    const destination = attemptRef.current?.destination;
    clearAttempt();
    if (destination) void begin(destination);
  };

  const cancel = () => {
    clearAttempt();
    setState(null);
    window.stop();
    void getDesktopBridge()?.cancelWorkspaceSwitch?.().catch(reportBridgeError);
  };

  return { state, switchWorkspace, retry, cancel };
};
