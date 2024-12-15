import { cn } from "lib";
import type { Story, StoryPriority } from "@/modules/stories/types";
import { StoriesGroup } from "@/components/ui/stories-group";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { BodyContainer } from "../shared/body";
import { useStatuses } from "@/lib/hooks/statuses";
import { StoriesList } from "./stories-list";
import { useMembers } from "@/lib/hooks/members";
import { Member } from "@/types";

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
  const { data: statuses = [] } = useStatuses();
  const { data: members = [] } = useMembers();

  const priorities: StoryPriority[] = [
    "Urgent",
    "High",
    "Medium",
    "Low",
    "No Priority",
  ];

  const unassignedMember: Member = {
    id: null,
    username: "Unassigned",
    email: "",
    fullName: "",
    avatarUrl: "",
    isActive: true,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  };

  return (
    <BodyContainer className={cn("overflow-x-auto pb-6", className)}>
      {groupBy === "None" && <StoriesList stories={stories} />}
      {groupBy === "Status" &&
        statuses.map((status) => (
          <StoriesGroup
            className="-top-[0.5px]"
            key={status.id}
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

      {groupBy === "Assignee" &&
        [...members, unassignedMember].map((member) => (
          <StoriesGroup
            className="-top-[0.5px]"
            key={member.id}
            assignee={member}
            stories={stories}
            viewOptions={viewOptions}
          />
        ))}
    </BodyContainer>
  );
};
