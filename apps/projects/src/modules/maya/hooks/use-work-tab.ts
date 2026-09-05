import { useCallback, useSyncExternalStore } from "react";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { MAYA_WORK_TABS, type MayaWorkTab } from "../utils/recent-work";

const CHANGED_EVENT = "maya-work-tab-changed";
const sessionTabs = new Map<string, MayaWorkTab>();
const getServerSnapshot = (): MayaWorkTab => "all";

const readTab = (key: string | null): MayaWorkTab => {
  if (!key) return "all";
  if (sessionTabs.has(key)) return sessionTabs.get(key)!;
  try {
    const value = window.localStorage.getItem(key);
    return MAYA_WORK_TABS.includes(value as MayaWorkTab)
      ? (value as MayaWorkTab)
      : "all";
  } catch {
    return "all";
  }
};

export const useWorkTab = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const key =
    session?.user.id && workspaceSlug
      ? `maya:work-tab:v1:${encodeURIComponent(session.user.id)}:${encodeURIComponent(workspaceSlug)}`
      : null;
  const subscribe = useCallback(
    (onChange: () => void) => {
      const onStorage = (event: StorageEvent) => {
        if (event.key === key || event.key === null) onChange();
      };
      const onTabChanged = (event: Event) => {
        if ((event as CustomEvent<string>).detail === key) onChange();
      };
      window.addEventListener("storage", onStorage);
      window.addEventListener(CHANGED_EVENT, onTabChanged);
      return () => {
        window.removeEventListener("storage", onStorage);
        window.removeEventListener(CHANGED_EVENT, onTabChanged);
      };
    },
    [key],
  );
  const getSnapshot = useCallback(() => readTab(key), [key]);
  const tab = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
  const setTab = useCallback(
    (value: MayaWorkTab) => {
      if (!key) return;
      try {
        window.localStorage.setItem(key, value);
        sessionTabs.delete(key);
      } catch {
        // Keep tab switching usable when browser storage is unavailable.
        sessionTabs.set(key, value);
      }
      window.dispatchEvent(new CustomEvent(CHANGED_EVENT, { detail: key }));
    },
    [key],
  );
  return [tab, setTab] as const;
};
