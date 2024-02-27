"use client";
import { BoardDividedPanel } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import type { IssueStatus, Issue } from "@/types/issue";
import { Header } from "./header";
import { Sidebar } from "./sidebar";
import { AllIssues } from "./all-issues";

export const ListMyIssues = ({
  issues,
  statuses,
}: {
  issues: Issue[];
  statuses: IssueStatus[];
}) => {
  const [isExpanded, setIsExpanded] = useLocalStorage(
    "my-issues:expanded",
    true,
  );
  return (
    <>
      <Header isExpanded={isExpanded} setIsExpanded={setIsExpanded} />
      <BoardDividedPanel autoSaveId="my-issues:divided-panel">
        <BoardDividedPanel.MainPanel>
          <AllIssues issues={issues} statuses={statuses} />
        </BoardDividedPanel.MainPanel>
        <BoardDividedPanel.SideBar isExpanded={isExpanded}>
          <Sidebar />
        </BoardDividedPanel.SideBar>
      </BoardDividedPanel>
    </>
  );
};
