import { cn } from "lib";
import { useParams } from "next/navigation";
import type {
  GroupedStoriesResponse,
  StoryPriority,
  StoryGroup,
  StoryFilters,
} from "@/modules/stories/types";
import { StoriesGroup } from "@/components/ui/stories-group";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { useStatuses, useTeamStatuses } from "@/lib/hooks/statuses";
import { useMembers } from "@/lib/hooks/members";
import type { Member } from "@/types";
import { useTeamMembers } from "@/lib/hooks/team-members";
import type { State } from "@/types/states";
import { BodyContainer } from "../shared/body";

// const unassignedMember: unknown = {
//   id: "unassigned",
//   username: "Unassigned",
//   email: "unassigned@example.com",
//   fullName: "",
//   avatarUrl: "",
//   isActive: true,
//   role: "member",
//   createdAt: new Date().toISOString(),
//   updatedAt: new Date().toISOString(),
// };

const GroupedStories = ({
  filters,
  group,
  groupBy,
  isInSearch,
  viewOptions,
  members,
  statuses,
  rowClassName,
}: {
  filters: StoryFilters;
  group: StoryGroup;
  groupBy: GroupedStoriesResponse["meta"]["groupBy"];
  isInSearch?: boolean;
  viewOptions: StoriesViewOptions;
  members: Member[];
  statuses: State[];
  rowClassName?: string;
}) => {
  const getGroupProps = () => {
    switch (groupBy) {
      case "priority":
        return { priority: group.key as StoryPriority };
      case "status":
        return { status: statuses.find((status) => status.id === group.key) };
      case "assignee":
        return { assignee: members.find((member) => member.id === group.key) };
      case "none":
        return {};
    }
  };

  return (
    <StoriesGroup
      className="-top-[0.5px]"
      id={group.key}
      isInSearch={isInSearch}
      key={group.key}
      {...getGroupProps()}
      filters={filters}
      group={group}
      rowClassName={rowClassName}
      viewOptions={viewOptions}
    />
  );
};

export const ListBoard = ({
  groupedStories,
  className,
  viewOptions,
  isInSearch,
  rowClassName,
}: {
  className?: string;
  groupedStories: GroupedStoriesResponse;
  viewOptions: StoriesViewOptions;
  isInSearch?: boolean;
  rowClassName?: string;
}) => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: allMembers = [] } = useMembers();
  const { data: teamMembers = [] } = useTeamMembers(teamId);
  const members = (teamId ? teamMembers : allMembers).filter(
    ({ role }) => role !== "system",
  );
  const { data: teamStatuses = [] } = useTeamStatuses(teamId);
  const { data: allStatuses = [] } = useStatuses();
  const statuses = teamId ? teamStatuses : allStatuses;

  return (
    <BodyContainer
      className={cn(
        "overflow-x-auto pb-6",
        {
          "h-auto pb-0": isInSearch,
        },
        className,
      )}
    >
      {groupedStories.groups.map((group) => (
        <GroupedStories
          filters={groupedStories.meta.filters}
          group={group}
          groupBy={groupedStories.meta.groupBy}
          isInSearch={isInSearch}
          key={group.key}
          members={members}
          rowClassName={rowClassName}
          statuses={statuses}
          viewOptions={viewOptions}
        />
      ))}
    </BodyContainer>
  );
};
