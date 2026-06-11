"use client";

import { useEffect } from "react";
import type { StoriesLayout } from "@/components/ui";
import { BoardDividedPanel } from "@/components/ui";
import { StoriesFilterBar } from "@/components/ui/stories-filter-bar";
import { useLocalStorage, useMediaQuery } from "@/hooks";
import { useChatContext } from "@/context/chat-context";
import { useSprint } from "../hooks/sprint-details";
import { Header } from "./header";
import { SprintStoriesProvider, useSprintOptions } from "./provider";
import { AllStories } from "./all-stories";
import { Sidebar } from "./sidebar";
import { StoriesSkeleton } from "./skeleton";

const ActiveStoriesFilterBar = () => {
  const { filters, resetFilters, setFilters } = useSprintOptions();

  return (
    <StoriesFilterBar
      filters={filters}
      hiddenFields={["teamIds", "sprintIds"]}
      resetFilters={resetFilters}
      setFilters={setFilters}
    />
  );
};

export const ListSprintStories = ({ sprintId }: { sprintId: string }) => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "team:sprints:stories:layout",
    "list",
  );
  const isMobile = useMediaQuery("(max-width: 768px)");
  const [isExpanded, setIsExpanded] = useLocalStorage(
    "team:sprints:stories:isExpanded",
    true,
  );
  const { isOpen: isChatOpen } = useChatContext();
  const { isPending: isSprintPending } = useSprint(sprintId);

  useEffect(() => {
    if (isChatOpen && isExpanded) {
      setIsExpanded(false);
    }
  }, [isChatOpen, isExpanded, setIsExpanded]);

  if (isSprintPending) {
    return <StoriesSkeleton layout={layout} />;
  }

  return (
    <SprintStoriesProvider layout={layout}>
      <Header
        isExpanded={isExpanded}
        layout={layout}
        setIsExpanded={setIsExpanded}
        setLayout={setLayout}
      />
      <ActiveStoriesFilterBar />

      {isMobile ? (
        <AllStories layout={layout} />
      ) : (
        <BoardDividedPanel autoSaveId="team:sprints:stories:divided-panel">
          <BoardDividedPanel.MainPanel>
            <AllStories layout={layout} />
          </BoardDividedPanel.MainPanel>
          <BoardDividedPanel.SideBar isExpanded={isExpanded}>
            <Sidebar />
          </BoardDividedPanel.SideBar>
        </BoardDividedPanel>
      )}
    </SprintStoriesProvider>
  );
};
