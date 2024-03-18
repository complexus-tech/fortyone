"use client";
import type { StoriesLayout } from "@/components/ui";
import { BoardDividedPanel } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import type { StoryStatus, Story } from "@/types/story";
import { Header } from "./header";
import { Sidebar } from "./sidebar";
import { AllStories } from "./all-stories";

export const ListMyStories = ({
  stories,
  statuses,
}: {
  stories: Story[];
  statuses: StoryStatus[];
}) => {
  const [isExpanded, setIsExpanded] = useLocalStorage(
    "my-stories:expanded",
    true,
  );
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "my-stories:stories:layout",
    "list",
  );
  return (
    <>
      <Header
        isExpanded={isExpanded}
        layout={layout}
        setIsExpanded={setIsExpanded}
        setLayout={setLayout}
      />
      <BoardDividedPanel autoSaveId="my-stories:divided-panel">
        <BoardDividedPanel.MainPanel>
          <AllStories layout={layout} statuses={statuses} stories={stories} />
        </BoardDividedPanel.MainPanel>
        <BoardDividedPanel.SideBar isExpanded={isExpanded}>
          <Sidebar />
        </BoardDividedPanel.SideBar>
      </BoardDividedPanel>
    </>
  );
};
