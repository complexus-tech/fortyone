import type { StoryAttachment } from "@/modules/story/types";
import type {
  DesktopEnqueueOperation,
  DesktopEnqueueResult,
  DesktopOutboxScope,
  DesktopOutboxSnapshot,
  DesktopResolveOperation,
} from "./outbox-types";

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

export type DesktopWorkspaceDestination = {
  id: string;
  name: string;
  url: string;
};

export type DesktopNotificationScope = {
  accountId: string;
  workspaceId: string;
  workspaceSlug: string;
};

export type DesktopNotification = DesktopNotificationScope & {
  notificationId: string;
  createdAt: string;
  title: string;
  body: string;
  entityType: string;
  entityId: string;
};

export type DesktopPreferences = {
  notifications: boolean;
  keepRunning: boolean;
  autostart: boolean;
};

export type DesktopCacheScope = {
  accountId: string;
  workspaceId: string;
  version: 1;
};

export type DesktopRestoredSession = DesktopNotificationScope & {
  expiresAt: number;
};

type DesktopBridge = {
  version: 2;
  platform: "macos" | "windows" | "linux";
  vibrancy?: boolean;
  syncChrome: (state: DesktopChromeState) => Promise<void>;
  navigationState: () => Promise<DesktopNavigationState>;
  navigate: (direction: DesktopNavigationDirection) => Promise<void>;
  windowAction: (action: DesktopWindowAction) => Promise<void>;
  // Optional capabilities keep the hosted app compatible with installed v2 builds.
  beginWorkspaceSwitch?: (
    destination: DesktopWorkspaceDestination,
  ) => Promise<void>;
  workspaceReady?: (workspace: { id: string }) => Promise<void>;
  cancelWorkspaceSwitch?: () => Promise<void>;
  getUpdateStatus?: () => Promise<{
    configured: boolean;
    currentVersion: string;
  }>;
  checkForUpdate?: () => Promise<{ available: boolean; version?: string }>;
  installUpdate?: (request: { version: string }) => Promise<void>;
  enqueueOperation?: (
    request: DesktopEnqueueOperation,
  ) => Promise<DesktopEnqueueResult>;
  readOutbox?: (scope: DesktopOutboxScope) => Promise<DesktopOutboxSnapshot>;
  syncOutbox?: (scope: DesktopOutboxScope) => Promise<DesktopOutboxSnapshot>;
  resolveOperation?: (
    request: DesktopResolveOperation,
  ) => Promise<DesktopOutboxSnapshot>;
  bindSession?: (scope: DesktopNotificationScope) => Promise<void>;
  restoreSession?: (scope: {
    workspaceSlug: string;
  }) => Promise<DesktopRestoredSession>;
  cacheRead?: (scope: DesktopCacheScope) => Promise<unknown>;
  cacheWrite?: (
    scope: DesktopCacheScope & { payload: unknown },
  ) => Promise<void>;
  cacheClear?: (scope: { accountId: string }) => Promise<void>;
  stageAttachment?: (attachment: {
    accountId: string;
    workspaceId: string;
    id: string;
    storyId: string;
    fileName: string;
    mimeType: string;
    data: string;
  }) => Promise<{
    operationId: string;
    attachment: StoryAttachment;
  }>;
  readAttachment?: (attachment: {
    accountId: string;
    workspaceId: string;
    storyId: string;
    attachmentId: string;
  }) => Promise<{ data: string; mimeType: string; fileName: string } | null>;
  exportAttachment?: (attachment: {
    accountId: string;
    workspaceId: string;
    storyId: string;
    attachmentId: string;
  }) => Promise<boolean>;
  exportRecovery?: (recovery: {
    accountId: string;
    workspaceId: string;
    documentId: string;
    payload: Record<string, unknown>;
  }) => Promise<boolean>;
  pageWrite?: (page: {
    accountId: string;
    workspaceId: string;
    url: string;
    html: string;
    contentSecurityPolicy?: string;
  }) => Promise<void>;
  pageRead?: (page: {
    accountId: string;
    workspaceId: string;
    url: string;
  }) => Promise<{
    html: string;
    savedAt: number;
    contentSecurityPolicy?: string;
  } | null>;
  draftRead?: (scope: {
    accountId: string;
    workspaceId: string;
    key: string;
  }) => Promise<unknown>;
  draftWrite?: (scope: {
    accountId: string;
    workspaceId: string;
    key: string;
    payload: unknown;
  }) => Promise<void>;
  draftDelete?: (scope: {
    accountId: string;
    workspaceId: string;
    key: string;
  }) => Promise<void>;
  getPreferences?: () => Promise<DesktopPreferences>;
  updatePreferences?: (
    preferences: Partial<DesktopPreferences>,
  ) => Promise<DesktopPreferences>;
  requestNotificationPermission?: () => Promise<
    "granted" | "denied" | "prompt"
  >;
  deliverNotification?: (notification: DesktopNotification) => Promise<void>;
  setNotificationBadge?: (
    state: DesktopNotificationScope & { count: number },
  ) => Promise<void>;
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
    ["macos", "windows", "linux"].includes(bridge.platform ?? "") &&
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
