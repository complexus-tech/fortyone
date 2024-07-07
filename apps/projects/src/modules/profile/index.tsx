"use client";
import type { StoriesLayout } from "@/components/ui";
import { BoardDividedPanel } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import type { Story } from "@/modules/stories/types";
import { Header } from "./components/header";
import { Sidebar } from "./components/sidebar";
import { AllStories } from "./components/all-stories";
import { ProfileProvider } from "./components/provider";

export const ListUserStories = ({
  stories,
  user,
}: {
  stories: Story[];
  user: string;
}) => {
  const [isExpanded, setIsExpanded] = useLocalStorage(
    `stories:${user}:expanded`,
    true,
  );
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    `stories:${user}:layout`,
    "list",
  );
  return (
    <ProfileProvider>
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
    </ProfileProvider>
  );
};
