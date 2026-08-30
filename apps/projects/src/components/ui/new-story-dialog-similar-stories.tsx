import type { ComponentProps } from "react";
import { SimilarItemsPanel } from "./similar-items-panel";
import { StoryRowPreview } from "./story/row";

type StoryPreviewAssignee = NonNullable<
  ComponentProps<typeof StoryRowPreview>["assignee"]
>;
type StoryPreviewPriority = ComponentProps<typeof StoryRowPreview>["priority"];

type SimilarStory = {
  assigneeId?: string | null;
  id: string;
  priority?: StoryPreviewPriority | null;
  sequenceId: number;
  statusId?: string | null;
  teamId: string;
  title: string;
};

type StoryStatus = {
  color: string;
  id: string;
  name: string;
};

export const NewStoryDialogSimilarStories = ({
  currentTeamCode,
  heading,
  members,
  onSelect,
  statuses,
  stories,
  teamCodes,
}: {
  currentTeamCode?: string;
  heading: string;
  members: (StoryPreviewAssignee & { id: string })[];
  onSelect: (story: {
    id: string;
    sequenceId: number;
    teamCode: string;
  }) => void;
  statuses: StoryStatus[];
  stories: SimilarStory[];
  teamCodes: { code: string; id: string }[];
}) => {
  const teamCodeById = new Map(teamCodes.map((team) => [team.id, team.code]));
  const statusById = new Map(statuses.map((status) => [status.id, status]));
  const memberById = new Map(members.map((member) => [member.id, member]));

  return (
    <SimilarItemsPanel heading={heading}>
      {stories.map((story) => {
        const teamCode =
          teamCodeById.get(story.teamId) ?? currentTeamCode ?? "";
        const status = story.statusId
          ? statusById.get(story.statusId)
          : undefined;
        const assignee = story.assigneeId
          ? memberById.get(story.assigneeId)
          : undefined;

        return (
          <StoryRowPreview
            assignee={assignee}
            key={story.id}
            onSelect={() => {
              onSelect({
                id: story.id,
                sequenceId: story.sequenceId,
                teamCode,
              });
            }}
            priority={story.priority ?? "No Priority"}
            reference={`${teamCode}-${story.sequenceId}`}
            statusColor={status?.color}
            statusId={story.statusId}
            statusName={status?.name}
            title={story.title}
          />
        );
      })}
    </SimilarItemsPanel>
  );
};
