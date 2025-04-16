"use client";

import type { StoriesLayout } from "@/components/ui";
import { BoardDividedPanel } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { useSprintStories } from "@/modules/stories/hooks/sprint-stories";
import { Header } from "./header";
import { SprintStoriesProvider } from "./provider";
import { AllStories } from "./all-stories";
import { Sidebar } from "./sidebar";
import { StoriesSkeleton } from "./skeleton";

export const ListSprintStories = ({ sprintId }: { sprintId: string }) => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "objective:sprints:layout",
    "list",
  );
  const [isExpanded, setIsExpanded] = useLocalStorage<boolean>(
    "objective:sprints:isExpanded",
    false,
  );
  const { isPending } = useSprintStories(sprintId);

  if (isPending) {
    return <StoriesSkeleton layout={layout} />;
  }

  return (
    <SprintStoriesProvider>
      <Header
        isExpanded={isExpanded}
        layout={layout}
        setIsExpanded={setIsExpanded}
        setLayout={setLayout}
      />
      <BoardDividedPanel autoSaveId="my-stories:divided-panel">
        <BoardDividedPanel.MainPanel>
          <AllStories layout={layout} />
        </BoardDividedPanel.MainPanel>
        <BoardDividedPanel.SideBar isExpanded={isExpanded}>
          <Sidebar />
        </BoardDividedPanel.SideBar>
      </BoardDividedPanel>
    </SprintStoriesProvider>
  );
};
