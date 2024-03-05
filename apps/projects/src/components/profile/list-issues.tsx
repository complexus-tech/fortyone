"use client";
import type { IssuesLayout } from "@/components/ui";
import { BoardDividedPanel } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import type { IssueStatus, Issue } from "@/types/issue";
import { Header } from "./header";
import { Sidebar } from "./sidebar";
import { AllIssues } from "./all-issues";

export const ListUserIssues = ({
  issues,
  statuses,
  user,
}: {
  issues: Issue[];
  statuses: IssueStatus[];
  user: string;
}) => {
  const [isExpanded, setIsExpanded] = useLocalStorage(
    `issues:${user}:expanded`,
    true,
  );
  const [layout, setLayout] = useLocalStorage<IssuesLayout>(
    `issues:${user}:layout`,
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
      <BoardDividedPanel autoSaveId="my-issues:divided-panel">
        <BoardDividedPanel.MainPanel>
          <AllIssues issues={issues} layout={layout} statuses={statuses} />
        </BoardDividedPanel.MainPanel>
        <BoardDividedPanel.SideBar isExpanded={isExpanded}>
          <Sidebar />
        </BoardDividedPanel.SideBar>
      </BoardDividedPanel>
    </>
  );
};
