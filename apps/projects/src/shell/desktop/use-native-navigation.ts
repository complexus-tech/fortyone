import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type {
  DesktopNavigationDirection,
  DesktopNavigationState,
} from "./bridge";
import { DESKTOP_HISTORY_EVENT, getDesktopBridge } from "./bridge";

const EMPTY_NAVIGATION: DesktopNavigationState = {
  canGoBack: false,
  canGoForward: false,
};

export const useNativeNavigation = (enabled: boolean) => {
  const [state, setState] = useState(EMPTY_NAVIGATION);
  const requestId = useRef(0);
  const frame = useRef<number | null>(null);
  const active = useRef(false);

  const refresh = useCallback(() => {
    if (!active.current) return;
    const currentRequest = ++requestId.current;
    if (frame.current !== null) cancelAnimationFrame(frame.current);

    // WKWebView updates its back-forward list after the history mutation.
    frame.current = requestAnimationFrame(() => {
      frame.current = null;
      const bridge = getDesktopBridge();
      if (!bridge) return;

      void bridge
        .navigationState()
        .then((nextState) => {
          if (currentRequest !== requestId.current) return;
          setState((previous) =>
            previous.canGoBack === nextState.canGoBack &&
            previous.canGoForward === nextState.canGoForward
              ? previous
              : nextState,
          );
        })
        .catch((error: unknown) => {
          if (currentRequest !== requestId.current) return;
          setState(EMPTY_NAVIGATION);
          // eslint-disable-next-line no-console -- Native failures must remain visible in desktop diagnostics.
          console.error("Could not read desktop navigation history", error);
        });
    });
  }, []);

  useEffect(() => {
    if (!enabled) return;
    active.current = true;
    const events = [
      DESKTOP_HISTORY_EVENT,
      "popstate",
      "hashchange",
      "pageshow",
    ];
    for (const event of events) window.addEventListener(event, refresh);
    refresh();

    return () => {
      active.current = false;
      requestId.current += 1;
      if (frame.current !== null) cancelAnimationFrame(frame.current);
      for (const event of events) window.removeEventListener(event, refresh);
    };
  }, [enabled, refresh]);

  const navigate = useCallback(
    (direction: DesktopNavigationDirection) => {
      const bridge = getDesktopBridge();
      if (!enabled || !bridge) return;

      void bridge
        .navigate(direction)
        .then(refresh)
        .catch((error: unknown) => {
          refresh();
          // eslint-disable-next-line no-console -- Native failures must remain visible in desktop diagnostics.
          console.error("Could not navigate desktop history", error);
        });
    },
    [enabled, refresh],
  );

  return useMemo(
    () => ({
      canGoBack: enabled && state.canGoBack,
      canGoForward: enabled && state.canGoForward,
      goBack: () => {
        if (state.canGoBack) navigate("back");
      },
      goForward: () => {
        if (state.canGoForward) navigate("forward");
      },
    }),
    [enabled, navigate, state],
  );
};
