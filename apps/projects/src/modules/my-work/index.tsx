"use client";
import type { StoriesLayout } from "@/components/ui";
import { BoardDividedPanel } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import type { Story } from "@/modules/stories/types";
import { Header } from "./components/header";
import { Sidebar } from "./components/sidebar";
import { ListStories } from "./components/list-stories";
import { MyWorkProvider } from "./components/provider";

export const ListMyStories = ({ stories }: { stories: Story[] }) => {
  const [isExpanded, setIsExpanded] = useLocalStorage(
    "my-stories:expanded",
    false,
  );
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "my-stories:stories:layout",
    "list",
  );
  return (
    <MyWorkProvider>
      <Header
        allStories={stories.length}
        isExpanded={isExpanded}
        layout={layout}
        setIsExpanded={setIsExpanded}
        setLayout={setLayout}
      />
      <BoardDividedPanel autoSaveId="my-stories:divided-panel">
        <BoardDividedPanel.MainPanel>
          <ListStories layout={layout} stories={stories} />
        </BoardDividedPanel.MainPanel>
        <BoardDividedPanel.SideBar isExpanded={isExpanded}>
          <Sidebar />
        </BoardDividedPanel.SideBar>
      </BoardDividedPanel>
    </MyWorkProvider>
  );
};
