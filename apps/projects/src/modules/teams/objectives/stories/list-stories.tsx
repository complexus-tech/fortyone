"use client";
import { useLocalStorage } from "@/hooks";
import type { StoriesLayout } from "@/components/ui";
import { BoardDividedPanel } from "@/components/ui";
import { Sidebar } from "./sidebar";
import { TeamOptionsProvider } from "./provider";
import { AllStories } from "./all-stories";
import { Header } from "../components/header";

export const ListStories = () => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "teams:stories:layout",
    "list",
  );
  const [isExpanded, setIsExpanded] = useLocalStorage(
    "teams:stories:expanded",
    false,
  );

  return (
    <TeamOptionsProvider>
      <Header
        isExpanded={isExpanded}
        layout={layout}
        setIsExpanded={setIsExpanded}
        setLayout={setLayout}
      />
      <BoardDividedPanel autoSaveId="teams:stories:divided-panel">
        <BoardDividedPanel.MainPanel>
          <AllStories layout={layout} />
        </BoardDividedPanel.MainPanel>
        <BoardDividedPanel.SideBar isExpanded={isExpanded}>
          <Sidebar />
        </BoardDividedPanel.SideBar>
      </BoardDividedPanel>
    </TeamOptionsProvider>
  );
};
