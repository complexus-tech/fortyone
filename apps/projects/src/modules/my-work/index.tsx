"use client";
import type { StoriesLayout } from "@/components/ui";
import { BoardDividedPanel } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { Header } from "./components/header";
import { Sidebar } from "./components/sidebar";
import { ListStories } from "./components/list-stories";
import { MyWorkProvider } from "./components/provider";

export const ListMyStories = () => {
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
          <ListStories layout={layout} />
        </BoardDividedPanel.MainPanel>
        <BoardDividedPanel.SideBar isExpanded={isExpanded}>
          <Sidebar />
        </BoardDividedPanel.SideBar>
      </BoardDividedPanel>
    </MyWorkProvider>
  );
};
