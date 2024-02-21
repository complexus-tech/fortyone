"use client";
import { Box } from "ui";
import { cn } from "lib";
import type { IssueStatus, Issue as IssueType } from "@/types/issue";
import { BodyContainer } from "@/components/layout";
import { Issue, IssuesHeader, IssuesToolbar } from "@/components/ui";

const ListIssues = ({ issues }: { issues: IssueType[] }) => {
  return (
    <>
      {issues.map((issue) => (
        <Issue issue={issue} key={issue.id} />
      ))}
    </>
  );
};

const IssuesGroup = ({
  issues,
  status,
}: {
  issues: IssueType[];
  status: IssueStatus;
}) => {
  const filteredIssues = issues.filter((issue) => issue.status === status);
  return (
    <Box
      className={cn({
        hidden: filteredIssues.length === 0,
      })}
    >
      <IssuesHeader count={filteredIssues.length} status={status} />
      <ListIssues issues={filteredIssues} />
    </Box>
  );
};

export const IssuesList = ({
  issues,
  statuses,
}: {
  issues: IssueType[];
  statuses: IssueStatus[];
}) => {
  return (
    <BodyContainer className="overflow-x-auto">
      {statuses.map((status) => (
        <IssuesGroup issues={issues} key={status} status={status} />
      ))}
      <IssuesToolbar />
    </BodyContainer>
  );
};
