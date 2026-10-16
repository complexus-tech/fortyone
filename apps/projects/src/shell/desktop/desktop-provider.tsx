"use client";

import type { ReactNode } from "react";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { useTheme } from "next-themes";
import type {
  DesktopAction,
  DesktopMenuState,
  DesktopShellState,
} from "./bridge";
import {
  DESKTOP_ACTION_EVENT,
  EMPTY_DESKTOP_SHELL,
  EMPTY_DESKTOP_MENU,
  getDesktopAction,
  getDesktopBridge,
  readDesktopBackground,
} from "./bridge";
import "./desktop.css";
import "./fallback.css";
import { DESKTOP_DRAG_HANDLERS, EMPTY_DRAG_HANDLERS } from "./drag-region";
import { DesktopHistoryControls } from "./desktop-history-controls";
import { useNativeNavigation } from "./use-native-navigation";
import { monitorDesktopConnectivity } from "./connectivity";

type DesktopContextValue = {
  isDesktop: boolean;
  setShellState: (state: DesktopShellState) => void;
  setShellMounted: (mounted: boolean) => void;
  registerMenu: (owner: string, state: DesktopMenuState) => void;
  unregisterMenu: (owner: string) => void;
  navigation: ReturnType<typeof useNativeNavigation>;
};

const DesktopContext = createContext<DesktopContextValue>({
  isDesktop: false,
  setShellState: () => {},
  setShellMounted: () => {},
  registerMenu: () => {},
  unregisterMenu: () => {},
  navigation: {
    canGoBack: false,
    canGoForward: false,
    goBack: () => {},
    goForward: () => {},
  },
});

export const DesktopProvider = ({ children }: { children: ReactNode }) => {
  const [isDesktop, setIsDesktop] = useState(false);
  const [shellState, setShellState] = useState(EMPTY_DESKTOP_SHELL);
  const [shellMounted, setShellMounted] = useState(false);
  const [menu, setMenu] = useState<{
    owner: string;
    state: DesktopMenuState;
  } | null>(null);
  const registerMenu = useCallback((owner: string, state: DesktopMenuState) => {
    setMenu({ owner, state });
  }, []);
  const unregisterMenu = useCallback((owner: string) => {
    setMenu((current) => (current?.owner === owner ? null : current));
  }, []);
  const menuSync = useRef<Promise<void> | null>(null);
  const resynchronizeChrome = useRef<(() => void) | null>(null);
  const { resolvedTheme, theme } = useTheme();
  const isMacDesktop = isDesktop && getDesktopBridge()?.platform === "macos";
  const navigation = useNativeNavigation(isMacDesktop);
  const showFallback = isMacDesktop && !shellMounted;

  useEffect(() => monitorDesktopConnectivity(), []);

  useEffect(() => {
    const bridge = getDesktopBridge();
    if (!bridge?.syncMenu) return;
    const synchronize = () => {
      // Serialize native writes so an outgoing layout cannot overwrite the
      // capabilities registered by the next layout while IPC is in flight.
      menuSync.current = (menuSync.current ?? Promise.resolve())
        .then(() => bridge.syncMenu!(menu?.state ?? EMPTY_DESKTOP_MENU))
        .catch((error: unknown) => {
          // eslint-disable-next-line no-console -- Keep native bridge failures visible in desktop diagnostics.
          console.error("Could not synchronize desktop menus", error);
        });
    };
    synchronize();
    window.addEventListener("pageshow", synchronize);
    return () => {
      window.removeEventListener("pageshow", synchronize);
    };
  }, [menu]);

  useEffect(() => {
    const bridge = getDesktopBridge();
    if (!bridge) return;

    let cancelled = false;
    let requestId = 0;
    let frame: number | null = null;
    const synchronize = () => {
      const currentRequest = ++requestId;
      if (frame !== null) cancelAnimationFrame(frame);
      frame = requestAnimationFrame(() => {
        frame = null;
        const background = readDesktopBackground();
        if (!background) return;

        void bridge
          .syncChrome({
            ...shellState,
            background,
            dark: resolvedTheme === "dark",
            systemTheme: !theme || theme === "system",
          })
          .then(() => {
            if (!cancelled && currentRequest === requestId) setIsDesktop(true);
          })
          .catch((error: unknown) => {
            if (cancelled || currentRequest !== requestId) return;
            // eslint-disable-next-line no-console -- Keep native bridge failures visible in desktop diagnostics.
            console.error("Could not synchronize the desktop toolbar", error);
          });
      });
    };
    resynchronizeChrome.current = synchronize;
    synchronize();

    return () => {
      cancelled = true;
      resynchronizeChrome.current = null;
      if (frame !== null) cancelAnimationFrame(frame);
    };
  }, [resolvedTheme, shellState, theme]);

  useEffect(() => {
    // A cached document retains React state after native page-load resets chrome.
    const handlePageShow = () => resynchronizeChrome.current?.();
    window.addEventListener("pageshow", handlePageShow);
    return () => {
      window.removeEventListener("pageshow", handlePageShow);
    };
  }, []);

  useEffect(() => {
    if (!isDesktop) return;
    const bridge = getDesktopBridge();
    document.documentElement.dataset.desktopPlatform = bridge?.platform;
    if (bridge?.platform === "macos" && bridge.vibrancy === true) {
      document.documentElement.dataset.desktopVibrancy = "true";
    }
    return () => {
      delete document.documentElement.dataset.desktopPlatform;
      delete document.documentElement.dataset.desktopVibrancy;
    };
  }, [isDesktop]);

  useEffect(() => {
    if (!showFallback) return;
    document.documentElement.dataset.desktopFallback = "true";
    return () => {
      delete document.documentElement.dataset.desktopFallback;
    };
  }, [showFallback]);

  const value = useMemo(
    () => ({
      isDesktop,
      setShellState,
      setShellMounted,
      registerMenu,
      unregisterMenu,
      navigation,
    }),
    [isDesktop, navigation, registerMenu, unregisterMenu],
  );

  return (
    <DesktopContext.Provider value={value}>
      {showFallback ? <DesktopFallbackTitlebar /> : null}
      {children}
    </DesktopContext.Provider>
  );
};

export const useDesktop = () => useContext(DesktopContext);
export const useDesktopNavigation = () => useDesktop().navigation;
export const useDesktopDragRegion = () =>
  useDesktop().isDesktop ? DESKTOP_DRAG_HANDLERS : EMPTY_DRAG_HANDLERS;

const DesktopFallbackTitlebar = () => {
  const navigation = useDesktopNavigation();
  const dragRegion = useDesktopDragRegion();

  return (
    <header
      aria-label="Desktop title bar"
      className="desktop-fallback-titlebar"
      {...dragRegion}
    >
      <DesktopHistoryControls {...navigation} />
    </header>
  );
};

export const useDesktopShell = ({
  ready,
  canCreateStory,
  storyLabel,
}: DesktopShellState) => {
  const { isDesktop, setShellState, setShellMounted } = useDesktop();

  useEffect(() => {
    setShellMounted(true);
    return () => {
      setShellMounted(false);
    };
  }, [setShellMounted]);

  useEffect(() => {
    if (!isDesktop) return;
    setShellState({ ready, canCreateStory, storyLabel });
  }, [canCreateStory, isDesktop, ready, setShellState, storyLabel]);

  useEffect(() => {
    if (!isDesktop) return;
    return () => {
      setShellState(EMPTY_DESKTOP_SHELL);
    };
  }, [isDesktop, setShellState]);

  return isDesktop && getDesktopBridge()?.platform === "macos";
};

export const useDesktopAction = (
  action: DesktopAction,
  onAction: () => void,
) => {
  const { isDesktop } = useDesktop();
  const handlerRef = useRef(onAction);
  useEffect(() => {
    handlerRef.current = onAction;
  }, [onAction]);

  useEffect(() => {
    if (!isDesktop) return;
    const handleAction = (event: Event) => {
      if (getDesktopAction(event) === action) handlerRef.current();
    };
    window.addEventListener(DESKTOP_ACTION_EVENT, handleAction);
    return () => {
      window.removeEventListener(DESKTOP_ACTION_EVENT, handleAction);
    };
  }, [action, isDesktop]);
};
