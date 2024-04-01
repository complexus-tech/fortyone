import { Box, Flex } from "ui";
import { cn } from "lib";
import type { Story, StoryPriority, StoryStatus } from "@/types/story";
import { BodyContainer } from "../shared/body";
import { StoriesKanbanHeader } from "./kanban-header";
import { KanbanGroup } from "./kanban-group";
import type { ViewOptionsGroupBy } from "./stories-view-options-button";

export const KanbanBoard = ({
  stories,
  className,
  groupBy,
}: {
  stories: Story[];
  className?: string;
  groupBy: ViewOptionsGroupBy;
}) => {
  const statuses: StoryStatus[] = [
    "Backlog",
    "Todo",
    "In Progress",
    "Testing",
    "Done",
    "Canceled",
  ];

  const priorities: StoryPriority[] = [
    "No Priority",
    "Low",
    "Medium",
    "High",
    "Urgent",
  ];

  return (
    <BodyContainer
      className={cn(
        "overflow-x-auto bg-gray-50/60 dark:bg-transparent",
        className,
      )}
    >
      <Box className="sticky top-0 z-[1] h-[3.5rem] w-max px-6 backdrop-blur">
        <Flex
          align="center"
          className="h-full shrink-0 overflow-x-auto"
          gap={6}
        >
          {groupBy === "Status" &&
            statuses.map((status) => (
              <StoriesKanbanHeader
                groupBy={groupBy}
                key={status}
                status={status}
                stories={stories}
              />
            ))}
          {groupBy === "Priority" &&
            priorities.map((priority) => (
              <StoriesKanbanHeader
                groupBy={groupBy}
                key={priority}
                priority={priority}
                stories={stories}
              />
            ))}
        </Flex>
      </Box>
      <Box className="flex h-[calc(100%-3.5rem)] w-max gap-x-6 px-7 ">
        {groupBy === "Status" &&
          statuses.map((status) => (
            <KanbanGroup
              groupBy={groupBy}
              key={status}
              status={status}
              stories={stories}
            />
          ))}
        {groupBy === "Priority" &&
          priorities.map((priority) => (
            <KanbanGroup
              groupBy={groupBy}
              key={priority}
              priority={priority}
              stories={stories}
            />
          ))}
      </Box>
    </BodyContainer>
  );
};
