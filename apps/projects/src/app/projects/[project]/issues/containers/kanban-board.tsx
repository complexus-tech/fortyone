import { Box, Flex } from "ui";
import type { Issue, IssueStatus } from "@/types/issue";
import { BodyContainer } from "@/components/layout";
import { KanbanList } from "../components/kanban-list";
import { Card } from "../components/card";
import { IssuesKanbanHeader } from "../components/kanban-header";

const IssuesGroup = ({
  issues,
  status,
}: {
  issues: Issue[];
  status: IssueStatus;
}) => {
  const filteredIssues = issues.filter((issue) => issue.status === status);
  return (
    <KanbanList key={status}>
      {filteredIssues.map((issue) => (
        <Card issue={issue} key={issue.id} />
      ))}
    </KanbanList>
  );
};

export const KanbanBoard = ({
  statuses,
  issues,
}: {
  statuses: IssueStatus[];
  issues: Issue[];
}) => {
  return (
    <BodyContainer className="overflow-x-auto bg-gray-50/60 dark:bg-transparent">
      <Box className="sticky top-0 z-[1] h-[3.5rem] w-max bg-gray-50/60 px-6 backdrop-blur dark:bg-dark-300/80">
        <Flex
          align="center"
          className="h-full shrink-0 overflow-x-auto"
          gap={6}
        >
          {statuses.map((status) => (
            <IssuesKanbanHeader issues={issues} key={status} status={status} />
          ))}
        </Flex>
      </Box>
      <Box className="flex w-max gap-x-6 px-7">
        {statuses.map((status) => (
          <IssuesGroup issues={issues} key={status} status={status} />
        ))}
      </Box>
    </BodyContainer>
  );
};
