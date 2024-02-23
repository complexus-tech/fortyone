import type { ReactNode } from "react";
import type { Issue, IssueStatus } from "@/types/issue";
import { IssueCard } from "./issue/card";

const List = ({ children }: { children: ReactNode }) => {
  return (
    <div className="flex h-[calc(100vh-7.5rem)] flex-col gap-3 overflow-y-auto pb-6">
      {children}
    </div>
  );
};

export const KanbanGroup = ({
  issues,
  status,
}: {
  issues: Issue[];
  status: IssueStatus;
}) => {
  const filteredIssues = issues.filter((issue) => issue.status === status);
  return (
    <List key={status}>
      {filteredIssues.map((issue) => (
        <IssueCard issue={issue} key={issue.id} />
      ))}
    </List>
  );
};
