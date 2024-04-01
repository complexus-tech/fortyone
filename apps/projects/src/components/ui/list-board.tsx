import { cn } from "lib";
import type { Story, StoryPriority, StoryStatus } from "@/types/story";
import { StoriesGroup } from "@/components/ui/stories-group";
import { BodyContainer } from "../shared/body";
import type { ViewOptionsGroupBy } from "./stories-view-options-button";

export const ListBoard = ({
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
    <BodyContainer className={cn("overflow-x-auto pb-6", className)}>
      {groupBy === "Status" &&
        statuses.map((status) => (
          <StoriesGroup
            className="-top-[0.5px]"
            groupBy={groupBy}
            key={status}
            status={status}
            stories={stories}
          />
        ))}
      {groupBy === "Priority" &&
        priorities.map((priority) => (
          <StoriesGroup
            className="-top-[0.5px]"
            groupBy={groupBy}
            key={priority}
            priority={priority}
            stories={stories}
          />
        ))}
    </BodyContainer>
  );
};
