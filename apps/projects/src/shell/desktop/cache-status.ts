export type DesktopCacheStatus = {
  workspaceId: string;
  phase: "restoring" | "settings" | "my-work" | "ready" | "paused";
  savedAt?: number;
  storageError?: string;
  downloadError?: string;
  shellError?: string;
  downloadedAt?: number;
};
let status: DesktopCacheStatus | null = null;
const listeners = new Set<() => void>();
export const getDesktopCacheStatus = () => status;
export const subscribeDesktopCacheStatus = (listener: () => void) => {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
};
export const updateDesktopCacheStatus = (
  next: Pick<DesktopCacheStatus, "workspaceId"> &
    Partial<Omit<DesktopCacheStatus, "workspaceId">>,
) => {
  status = {
    phase: "restoring",
    ...(status?.workspaceId === next.workspaceId ? status : {}),
    ...next,
  };
  for (const listener of listeners) listener();
};
export const DESKTOP_CACHE_RETRY_EVENT = "fortyone:desktop-cache-retry";

export const DESKTOP_CACHE_SAVE_EVENT = "fortyone:desktop-cache-save";
