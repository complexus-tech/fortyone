"use client";

import type { ComponentType, ReactNode } from "react";
import { createContext, createElement, useContext } from "react";
import type { SavedViewConfiguration } from "@/shared/story/view-configuration";

export type SaveViewProps = {
  teamId?: string;
  configuration: SavedViewConfiguration;
};
const SaveViewContext = createContext<ComponentType<SaveViewProps> | null>(
  null,
);

export const SaveViewProvider = ({
  children,
  SaveView,
}: {
  children: ReactNode;
  SaveView: ComponentType<SaveViewProps>;
}) => (
  <SaveViewContext.Provider value={SaveView}>
    {children}
  </SaveViewContext.Provider>
);

export const SaveViewSlot = (props: SaveViewProps) => {
  const SaveView = useContext(SaveViewContext);
  return SaveView ? createElement(SaveView, props) : null;
};
