"use client";
import { useLocalStorage } from "@/hooks";
import type { StoriesLayout } from "@/components/ui";
import { BoardDividedPanel } from "@/components/ui";
import { Sidebar } from "./sidebar";
import { ObjectiveOptionsProvider } from "./provider";
import { AllStories } from "./all-stories";
import { Header } from "./header";

export const ListStories = () => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "teams:objectives:stories:layout",
    "list",
  );
  const [isExpanded, setIsExpanded] = useLocalStorage(
    "teams:objectives:stories:expanded",
    false,
  );

  return (
    <ObjectiveOptionsProvider>
      <Header
        isExpanded={isExpanded}
        layout={layout}
        setIsExpanded={setIsExpanded}
        setLayout={setLayout}
      />
      <BoardDividedPanel autoSaveId="teams:objectives:stories:divided-panel">
        <BoardDividedPanel.MainPanel>
          <AllStories layout={layout} />
        </BoardDividedPanel.MainPanel>
        <BoardDividedPanel.SideBar isExpanded={isExpanded}>
          <Sidebar />
        </BoardDividedPanel.SideBar>
      </BoardDividedPanel>
    </ObjectiveOptionsProvider>
  );
};
