"use client";
import type { ReactNode } from "react";
import { createContext, useContext } from "react";
import type { DesktopCacheScope } from "./cache-types";

const Scope = createContext<DesktopCacheScope | null>(null);
export const DesktopWorkspaceScope = ({
  scope,
  children,
}: {
  scope: DesktopCacheScope | null;
  children: ReactNode;
}) => <Scope.Provider value={scope}>{children}</Scope.Provider>;
export const useDesktopWorkspaceScope = () => useContext(Scope);
