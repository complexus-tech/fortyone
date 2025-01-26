"use client";
import type { StoriesLayout } from "@/components/ui";
import { BoardDividedPanel } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { Header } from "./header";
import { SprintStoriesProvider } from "./provider";
import { AllStories } from "./all-stories";
import { Sidebar } from "./sidebar";

export const ListSprintStories = () => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "objective:sprints:layout",
    "list",
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
