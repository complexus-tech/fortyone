import { Box, Tabs } from "ui";
import type { IssuesLayout } from "@/components/ui";
import { IssuesBoard } from "@/components/ui";
import type { IssueStatus, Issue } from "@/types/issue";

export const AllIssues = ({
  issues,
  statuses,
  layout,
}: {
  issues: Issue[];
  statuses: IssueStatus[];
  layout: IssuesLayout;
}) => {
  return (
    <Box className="h-[calc(100vh-4rem)]">
      <Tabs defaultValue="assigned">
        <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b border-gray-100 backdrop-blur-lg dark:border-dark-100/40">
          <Tabs.List>
            <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
            <Tabs.Tab value="created">Created</Tabs.Tab>
            <Tabs.Tab value="subscribed">Subscribed</Tabs.Tab>
          </Tabs.List>
        </Box>
        <Tabs.Panel value="assigned">
          <IssuesBoard
            className="h-[calc(100vh-7.7rem)]"
            issues={issues}
            layout={layout}
            statuses={statuses}
          />
        </Tabs.Panel>
        <Tabs.Panel value="created">
          <IssuesBoard
            className="h-[calc(100vh-7.7rem)]"
            issues={issues}
            layout={layout}
            statuses={statuses}
          />
        </Tabs.Panel>
        <Tabs.Panel value="subscribed">
          <IssuesBoard
            className="h-[calc(100vh-7.7rem)]"
            issues={issues}
            layout={layout}
            statuses={statuses}
          />
        </Tabs.Panel>
      </Tabs>
    </Box>
  );
};
