"use client";
import { Suspense } from "react";
import type { StoriesLayout } from "@/components/ui";
import { BoardDividedPanel } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { BoardSkeleton } from "@/components/ui/board-skeleton";
import { Header } from "./header";
import { SprintStoriesProvider } from "./provider";
import { AllStories } from "./all-stories";
import { Sidebar } from "./sidebar";

export const ListSprintStories = () => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "objective:sprints:layout",
    "kanban",
  );
  const [isExpanded, setIsExpanded] = useLocalStorage<boolean>(
    "objective:sprints:isExpanded",
    false,
  );

  return (
    <SprintStoriesProvider>
      <Header
        isExpanded={isExpanded}
        layout={layout}
        setIsExpanded={setIsExpanded}
        setLayout={setLayout}
      />
      <Suspense fallback={<BoardSkeleton layout={layout} />}>
        <BoardDividedPanel autoSaveId="my-stories:divided-panel">
          <BoardDividedPanel.MainPanel>
            <AllStories layout={layout} />
          </BoardDividedPanel.MainPanel>
          <BoardDividedPanel.SideBar isExpanded={isExpanded}>
            <Sidebar />
          </BoardDividedPanel.SideBar>
        </BoardDividedPanel>
      </Suspense>
    </SprintStoriesProvider>
  );
};
