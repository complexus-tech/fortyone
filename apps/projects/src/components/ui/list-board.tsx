import { cn } from "lib";
import type {
  Story,
  StoryPriority,
  StoryStatus,
} from "@/modules/stories/types";
import { StoriesGroup } from "@/components/ui/stories-group";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { BodyContainer } from "../shared/body";

export const ListBoard = ({
  stories,
  className,
  viewOptions,
}: {
  stories: Story[];
  className?: string;
  viewOptions: StoriesViewOptions;
}) => {
  const { groupBy } = viewOptions;
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
            key={status}
            status={status}
            stories={stories}
            viewOptions={viewOptions}
          />
        ))}
      {groupBy === "Priority" &&
        priorities.map((priority) => (
          <StoriesGroup
            className="-top-[0.5px]"
            key={priority}
            priority={priority}
            stories={stories}
            viewOptions={viewOptions}
          />
        ))}
    </BodyContainer>
  );
};
