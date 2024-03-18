import { Box, Flex } from "ui";
import { cn } from "lib";
import type { Story, StoryStatus } from "@/types/story";
import { BodyContainer } from "../shared/body";
import { StoriesKanbanHeader } from "./kanban-header";
import { KanbanGroup } from "./kanban-group";

export const KanbanBoard = ({
  statuses,
  stories,
  className,
}: {
  statuses: StoryStatus[];
  stories: Story[];
  className?: string;
}) => {
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
          {statuses.map((status) => (
            <StoriesKanbanHeader
              stories={stories}
              key={status}
              status={status}
            />
          ))}
        </Flex>
      </Box>
      <Box className="flex h-[calc(100%-3.5rem)] w-max gap-x-6 px-7 ">
        {statuses.map((status) => (
          <KanbanGroup stories={stories} key={status} status={status} />
        ))}
      </Box>
    </BodyContainer>
  );
};
