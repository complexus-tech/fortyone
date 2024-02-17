"use client";
import { Box, Tabs } from "ui";
import { cn } from "lib";
import type { IssueStatus, Issue as IssueType } from "@/types/issue";
import { BodyContainer } from "@/components/layout";
import { Issue, IssuesHeader, IssuesToolbar } from "@/components/ui";

const ListIssues = ({ issues }: { issues: IssueType[] }) => {
  return (
    <>
      {issues.map(({ id, status, priority, title }) => (
        <Issue key={id} priority={priority} status={status} title={title} />
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
    <BodyContainer className="relative">
      <Tabs defaultValue="assigned">
        <Box className="sticky top-0 z-10 border-b border-gray-100 py-3 backdrop-blur dark:border-dark-200 dark:bg-dark-300/80">
          <Tabs.List>
            <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
            <Tabs.Tab value="created">Created</Tabs.Tab>
            <Tabs.Tab value="subscribed">Subscribed</Tabs.Tab>
          </Tabs.List>
        </Box>
        <Tabs.Panel value="assigned">
          {statuses.map((status) => (
            <IssuesGroup issues={issues} key={status} status={status} />
          ))}
        </Tabs.Panel>
        <Tabs.Panel value="created">
          <ListIssues issues={issues} />
        </Tabs.Panel>
        <Tabs.Panel value="subscribed">
          <ListIssues issues={issues} />
        </Tabs.Panel>
      </Tabs>
      <IssuesToolbar />
    </BodyContainer>
  );
};
