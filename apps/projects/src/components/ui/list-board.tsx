import { cn } from "lib";
import { useParams } from "next/navigation";
import type {
  GroupedStoriesResponse,
  StoryPriority,
  StoryGroup,
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
  group,
  groupBy,
  isInSearch,
  viewOptions,
  members,
  statuses,
}: {
  group: StoryGroup;
  groupBy: GroupedStoriesResponse["meta"]["groupBy"];
  isInSearch?: boolean;
  viewOptions: StoriesViewOptions;
  members: Member[];
  statuses: State[];
}) => {
  const getGroupProps = () => {
    switch (groupBy) {
      case "priority":
        return { priority: group.key as StoryPriority };
      case "status":
        return { status: statuses.find((status) => status.id === group.key) };
      case "assignee":
        return { assignee: members.find((member) => member.id === group.key) };
    }
  };

  return (
    <StoriesGroup
      className="-top-[0.5px]"
      isInSearch={isInSearch}
      key={group.key}
      {...getGroupProps()}
      stories={group.stories}
      viewOptions={viewOptions}
    />
  );
};

export const ListBoard = ({
  groupedStories,
  className,
  viewOptions,
  isInSearch,
}: {
  className?: string;
  groupedStories: GroupedStoriesResponse;
  viewOptions: StoriesViewOptions;
  isInSearch?: boolean;
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
          group={group}
          groupBy={groupedStories.meta.groupBy}
          isInSearch={isInSearch}
          key={group.key}
          members={members}
          statuses={statuses}
          viewOptions={viewOptions}
        />
      ))}
      {/* {groupBy === "none" && (
        <StoriesList isInSearch={isInSearch} stories={stories} />
      )} */}
    </BodyContainer>
  );
};
