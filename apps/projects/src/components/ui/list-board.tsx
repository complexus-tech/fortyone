import { cn } from "lib";
import type { Story, StoryPriority } from "@/modules/stories/types";
import { StoriesGroup } from "@/components/ui/stories-group";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { useStatuses } from "@/lib/hooks/statuses";
import { useMembers } from "@/lib/hooks/members";
import type { Member } from "@/types";
import { BodyContainer } from "../shared/body";
import { StoriesList } from "./stories-list";

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

  const unassignedMember: unknown = {
    id: "unassigned",
    username: "Unassigned",
    email: "unassigned@example.com",
    fullName: "",
    avatarUrl: "",
    isActive: true,
    role: "member",
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
            assignee={member as Member}
            className="-top-[0.5px]"
            key={(member as Member).email}
            stories={stories}
            viewOptions={viewOptions}
          />
        ))}
    </BodyContainer>
  );
};
