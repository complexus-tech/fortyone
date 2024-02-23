import { Box, Flex } from "ui";
import type { Issue, IssueStatus } from "@/types/issue";
import { BodyContainer } from "../layout/body";
import { IssuesKanbanHeader } from "./kanban-header";
import { KanbanGroup } from "./kanban-group";

export const KanbanBoard = ({
  statuses,
  issues,
}: {
  statuses: IssueStatus[];
  issues: Issue[];
}) => {
  return (
    <BodyContainer className="overflow-x-auto bg-gray-50/60 dark:bg-transparent">
      <Box className="sticky top-0 z-[1] h-[3.5rem] w-max px-6 backdrop-blur">
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
          <KanbanGroup issues={issues} key={status} status={status} />
        ))}
      </Box>
    </BodyContainer>
  );
};
