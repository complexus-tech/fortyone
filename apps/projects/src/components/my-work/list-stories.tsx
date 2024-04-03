"use client";
import type { StoriesLayout } from "@/components/ui";
import { BoardDividedPanel } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import type { Story } from "@/types/story";
import { Header } from "./header";
import { Sidebar } from "./sidebar";
import { AllStories } from "./all-stories";
import { MyWorkProvider } from "./provider";

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
        isExpanded={isExpanded}
        layout={layout}
        setIsExpanded={setIsExpanded}
        setLayout={setLayout}
      />
      <BoardDividedPanel autoSaveId="my-stories:divided-panel">
        <BoardDividedPanel.MainPanel>
          <AllStories layout={layout} stories={stories} />
        </BoardDividedPanel.MainPanel>
        <BoardDividedPanel.SideBar isExpanded={isExpanded}>
          <Sidebar />
        </BoardDividedPanel.SideBar>
      </BoardDividedPanel>
    </MyWorkProvider>
  );
};
