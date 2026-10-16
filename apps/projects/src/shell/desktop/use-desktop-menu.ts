"use client";

import { useEffect, useId } from "react";
import type { DesktopMenuState } from "./bridge";
import { useDesktop } from "./desktop-provider";

export const useDesktopMenu = ({
  ready,
  canCreateStory,
  canCreateObjective,
  canToggleSidebar,
  sidebarCollapsed,
  storyLabel,
  objectiveLabel,
}: DesktopMenuState) => {
  const { isDesktop, registerMenu, unregisterMenu } = useDesktop();
  const owner = useId();

  useEffect(() => {
    if (!isDesktop) return;
    registerMenu(owner, {
      ready,
      canCreateStory,
      canCreateObjective,
      canToggleSidebar,
      sidebarCollapsed,
      storyLabel,
      objectiveLabel,
    });
  }, [
    isDesktop,
    owner,
    registerMenu,
    ready,
    canCreateStory,
    canCreateObjective,
    canToggleSidebar,
    sidebarCollapsed,
    storyLabel,
    objectiveLabel,
  ]);

  useEffect(() => {
    return () => {
      unregisterMenu(owner);
    };
  }, [owner, unregisterMenu]);
};
