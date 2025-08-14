"use client";

import type { StoriesLayout } from "@/components/ui";
import { BoardDividedPanel } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { useSprint } from "../hooks/sprint-details";
import { Header } from "./header";
import { SprintStoriesProvider } from "./provider";
import { AllStories } from "./all-stories";
import { Sidebar } from "./sidebar";
import { StoriesSkeleton } from "./skeleton";

export const ListSprintStories = ({ sprintId }: { sprintId: string }) => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "team:sprints:stories:layout",
    "list",
  );
  const [isExpanded, setIsExpanded] = useLocalStorage(
    "team:sprints:stories:isExpanded",
    true,
  );
  const { isPending: isSprintPending } = useSprint(sprintId);

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
      <BoardDividedPanel autoSaveId="team:sprints:stories:divided-panel">
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
