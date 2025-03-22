"use client";
import { useLocalStorage } from "@/hooks";
import type { StoriesLayout } from "@/components/ui";
import { BoardDividedPanel } from "@/components/ui";
import { useTeamStories } from "@/modules/stories/hooks/team-stories";
import { Sidebar } from "./sidebar";
import { TeamOptionsProvider } from "./provider";
import { Header } from "./header";
import { AllStories } from "./all-stories";
import { StoriesSkeleton } from "./stories-skeleton";

export const ListStories = ({ teamId }: { teamId: string }) => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "teams:stories:layout",
    "kanban",
  );
  const [isExpanded, setIsExpanded] = useLocalStorage(
    "teams:stories:expanded",
    false,
  );
  const { isPending } = useTeamStories(teamId);
  if (isPending) {
    return <StoriesSkeleton layout={layout} />;
  }

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
