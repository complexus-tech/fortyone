import { Box } from "ui";
import { cn } from "lib";
import type { IssueStatus, Issue } from "@/types/issue";
import { IssuesHeader } from "./issues-header";
import { IssuesList } from "./issues-list";

export const IssuesGroup = ({
  issues,
  status,
  className,
}: {
  issues: Issue[];
  status: IssueStatus;
  className?: string;
}) => {
  const filteredIssues = issues.filter((issue) => issue.status === status);

  return (
    <Box
      className={cn("pb-6", {
        hidden: filteredIssues.length === 0,
      })}
    >
      <IssuesHeader
        className={className}
        count={filteredIssues.length}
        status={status}
      />
      <IssuesList id={status} issues={filteredIssues} />
    </Box>
  );
};
