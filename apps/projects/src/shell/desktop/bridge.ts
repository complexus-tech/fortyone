export const DESKTOP_ACTION_EVENT = "fortyone:desktop-action";
export const DESKTOP_HISTORY_EVENT = "fortyone:desktop-history-change";

export type DesktopAction = "create-story" | "search";

export type DesktopShellState = {
  ready: boolean;
  canCreateStory: boolean;
  storyLabel: string;
};

export type DesktopChromeState = DesktopShellState & {
  background: [number, number, number];
  dark: boolean;
  systemTheme: boolean;
};

export type DesktopNavigationState = {
  canGoBack: boolean;
  canGoForward: boolean;
};

export type DesktopNavigationDirection = "back" | "forward";
export type DesktopWindowAction = "drag" | "toggle-maximize";

type DesktopBridge = {
  version: 2;
  platform: "macos";
  vibrancy?: boolean;
  syncChrome: (state: DesktopChromeState) => Promise<void>;
  navigationState: () => Promise<DesktopNavigationState>;
  navigate: (direction: DesktopNavigationDirection) => Promise<void>;
  windowAction: (action: DesktopWindowAction) => Promise<void>;
};

declare global {
  interface Window {
    __FORTYONE_DESKTOP__?: DesktopBridge;
  }
}

export const EMPTY_DESKTOP_SHELL: DesktopShellState = {
  ready: false,
  canCreateStory: false,
  storyLabel: "story",
};

export const getDesktopBridge = () => {
  if (typeof window === "undefined") return null;

  // The injected JavaScript may come from an older installed desktop build.
  const bridge = window.__FORTYONE_DESKTOP__ as
    | Partial<DesktopBridge>
    | undefined;
  return bridge?.version === 2 &&
    bridge.platform === "macos" &&
    typeof bridge.syncChrome === "function" &&
    typeof bridge.navigationState === "function" &&
    typeof bridge.navigate === "function" &&
    typeof bridge.windowAction === "function"
    ? (bridge as DesktopBridge)
    : null;
};

export const getDesktopAction = (event: Event): DesktopAction | null => {
  if (!(event instanceof CustomEvent)) return null;
  const detail: unknown = event.detail;
  if (!detail || typeof detail !== "object" || !("action" in detail)) {
    return null;
  }

  const { action } = detail;
  return action === "create-story" || action === "search" ? action : null;
};

export const readDesktopBackground = (): [number, number, number] | null => {
  // The desktop body is translucent; resolve the opaque theme token so theme
  // changes and restored pages can still synchronize the native window.
  const color = getComputedStyle(document.documentElement)
    .getPropertyValue("--color-background")
    .trim();
  if (!color) return null;
  const canvas = document.createElement("canvas");
  canvas.width = 1;
  canvas.height = 1;
  const context = canvas.getContext("2d", { willReadFrequently: true });
  if (!context) return null;

  // Canvas resolves the shared OKLCH theme tokens into the native sRGB space.
  context.fillStyle = color;
  context.fillRect(0, 0, 1, 1);
  const { data } = context.getImageData(0, 0, 1, 1);
  if (data[3] === 0) return null;
  return [data[0], data[1], data[2]];
};
