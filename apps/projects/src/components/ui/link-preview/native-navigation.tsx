"use client";

import type { MouseEvent, ReactNode } from "react";
import { createContext, useContext } from "react";
import type { NativeLinkTarget } from "./native-target";

export type NativeLinkNavigation = (
  target: NativeLinkTarget,
  trigger: HTMLElement,
) => boolean;

const NativeNavigationContext = createContext<NativeLinkNavigation | null>(
  null,
);

export const NativeLinkNavigationProvider = ({
  children,
  onOpen,
}: {
  children: ReactNode;
  onOpen: NativeLinkNavigation;
}) => (
  <NativeNavigationContext.Provider value={onOpen}>
    {children}
  </NativeNavigationContext.Provider>
);

export const useNativeLinkNavigation = () =>
  useContext(NativeNavigationContext);

/** Preserve open-in-new-tab and browser context-menu behavior. */
export const openNativeLink = (
  event: MouseEvent<HTMLAnchorElement>,
  target: NativeLinkTarget | null,
  navigate: NativeLinkNavigation | null,
) => {
  if (
    !target ||
    !navigate ||
    event.defaultPrevented ||
    event.button !== 0 ||
    event.metaKey ||
    event.ctrlKey ||
    event.shiftKey ||
    event.altKey
  )
    return;
  if (navigate(target, event.currentTarget)) event.preventDefault();
};
