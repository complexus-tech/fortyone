import { Box, Flex } from "ui";
import { cn } from "lib";
import type { Story, StoryPriority } from "@/modules/stories/types";
import { BodyContainer } from "../shared/body";
import { StoriesKanbanHeader } from "./kanban-header";
import { KanbanGroup } from "./kanban-group";
import { useBoard } from "./board-context";
import { useStore } from "@/hooks/store";

export const KanbanBoard = ({
  stories,
  className,
}: {
  stories: Story[];
  className?: string;
}) => {
  const { viewOptions } = useBoard();
  const { groupBy } = viewOptions;
  const { states: statuses } = useStore();

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
                key={status.id}
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
      <Box className="flex h-[calc(100%-3.5rem)] w-max gap-x-6 px-7">
        {groupBy === "Status" &&
          statuses.map((status) => (
            <KanbanGroup
              groupBy={groupBy}
              key={status.id}
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
