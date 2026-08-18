"use client";

import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  type ReactNode,
} from "react";
import { useHotkeys } from "react-hotkeys-hook";
import { useLocalStorage, useWorkspacePath } from "@/hooks";

type SidebarContextValue = {
  isCollapsed: boolean;
  setIsCollapsed: (
    value: boolean | ((currentValue: boolean) => boolean),
  ) => void;
  toggleSidebar: () => void;
};

const SidebarContext = createContext<SidebarContextValue | null>(null);

export const SidebarProvider = ({ children }: { children: ReactNode }) => {
  const { workspaceSlug } = useWorkspacePath();
  const [isCollapsed, setIsCollapsed] = useLocalStorage(
    `sidebar:${workspaceSlug}:collapsed`,
    false,
  );
  const toggleSidebar = useCallback(() => {
    setIsCollapsed((currentValue) => !currentValue);
  }, [setIsCollapsed]);

  useHotkeys("mod+b", toggleSidebar, { preventDefault: true });

  const value = useMemo(
    () => ({ isCollapsed, setIsCollapsed, toggleSidebar }),
    [isCollapsed, setIsCollapsed, toggleSidebar],
  );

  return (
    <SidebarContext.Provider value={value}>{children}</SidebarContext.Provider>
  );
};

export const useSidebar = () => {
  const context = useContext(SidebarContext);

  if (!context) {
    throw new Error("useSidebar must be used within SidebarProvider");
  }

  return context;
};
