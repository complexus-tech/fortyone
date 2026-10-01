"use client";

import type { ComponentType, ReactNode } from "react";
import { createContext, useContext } from "react";

export type BoardPropertyStory = {
  id: string;
  teamId: string;
  subStories?: BoardPropertyStory[];
};

export type BoardPropertyProviderProps = {
  children: ReactNode;
  stories: BoardPropertyStory[];
  selectedIds?: string[];
};

export type StoryPropertyBadgesProps = {
  storyId: string;
  teamId: string;
  asList?: boolean;
  disabled?: boolean;
};

export type StoryPropertyDisplayPickerProps = {
  selectedIds: string[];
  onChange: (ids: string[]) => void;
};

export type WorkflowCountsProviderProps = {
  teamId?: string;
  enabled: boolean;
  children: (counts: ReadonlyMap<string, number>) => ReactNode;
};

export type BoardPropertySlots = {
  Provider: ComponentType<BoardPropertyProviderProps>;
  Badges: ComponentType<StoryPropertyBadgesProps>;
  DisplayPicker: ComponentType<StoryPropertyDisplayPickerProps>;
  WorkflowCounts: ComponentType<WorkflowCountsProviderProps>;
};

const PassthroughProvider = ({ children }: BoardPropertyProviderProps) => (
  <>{children}</>
);
const EmptyPropertySlot = () => null;
const NO_WORKFLOW_COUNTS: ReadonlyMap<string, number> = new Map();
const EmptyWorkflowCountsProvider = ({
  children,
}: WorkflowCountsProviderProps) => <>{children(NO_WORKFLOW_COUNTS)}</>;

const DEFAULT_SLOTS: BoardPropertySlots = {
  Provider: PassthroughProvider,
  Badges: EmptyPropertySlot,
  DisplayPicker: EmptyPropertySlot,
  WorkflowCounts: EmptyWorkflowCountsProvider,
};

const BoardPropertySlotsContext = createContext(DEFAULT_SLOTS);

/** The shell composes feature components; shared task UI only knows their contract. */
export const BoardPropertySlotsProvider = ({
  children,
  slots,
}: {
  children: ReactNode;
  slots: BoardPropertySlots;
}) => (
  <BoardPropertySlotsContext.Provider value={slots}>
    {children}
  </BoardPropertySlotsContext.Provider>
);

export const useBoardPropertySlots = () =>
  useContext(BoardPropertySlotsContext);
