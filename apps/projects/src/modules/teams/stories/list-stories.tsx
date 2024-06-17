"use client";
import type { Story } from "@/types/story";
import { useLocalStorage } from "@/hooks";
import type { StoriesLayout } from "@/components/ui";
import { BoardDividedPanel } from "@/components/ui";
import { Sidebar } from "./sidebar";
import { TeamStoriesProvider } from "./provider";
import { Header } from "./header";
import { AllStories } from "./all-stories";

export const ListStories = ({ stories }: { stories: Story[] }) => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "teams:stories:layout",
    "list",
  );
  const [isExpanded, setIsExpanded] = useLocalStorage(
    "teams:stories:expanded",
    false,
  );

  return (
    <TeamStoriesProvider>
      <Header
        allStories={stories.length}
        isExpanded={isExpanded}
        layout={layout}
        setIsExpanded={setIsExpanded}
        setLayout={setLayout}
      />
      <BoardDividedPanel autoSaveId="teams:stories:divided-panel">
        <BoardDividedPanel.MainPanel>
          <AllStories layout={layout} stories={stories} />
        </BoardDividedPanel.MainPanel>
        <BoardDividedPanel.SideBar isExpanded={isExpanded}>
          <Sidebar stories={stories} />
        </BoardDividedPanel.SideBar>
      </BoardDividedPanel>
    </TeamStoriesProvider>
  );
};
