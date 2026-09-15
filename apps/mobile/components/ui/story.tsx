import type { Story as StoryType } from "@/modules/stories/types";
import type { DisplayColumn } from "@/types/stories-view-options";
import { useTeamStatuses } from "@/modules/statuses";
import { useMembers } from "@/modules/members";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { StoryRow } from "@/modules/stories/components/story-row";
import { getCompletionStatus } from "@/modules/stories/utils/quick-actions";

// Standalone consumers resolve metadata here; grouped lists pass their shared
// lookup results directly to the same presentation component.
export const Story = ({
  story,
  visibleColumns = ["Status", "Assignee", "Priority"],
}: {
  story: StoryType;
  visibleColumns?: DisplayColumn[];
}) => {
  const { data: statuses = [] } = useTeamStatuses(story.teamId);
  const { data: members = [] } = useMembers();
  const { data: teams = [] } = useTeams();
  return (
    <StoryRow
      story={story}
      visibleColumns={visibleColumns}
      status={statuses.find((status) => status.id === story.statusId)}
      completionStatus={getCompletionStatus(statuses, story.teamId)}
      assignee={members.find((member) => member.id === story.assigneeId)}
      team={teams.find((team) => team.id === story.teamId)}
    />
  );
};
