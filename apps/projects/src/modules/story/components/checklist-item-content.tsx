import Link from "next/link";
import { Badge, Tooltip } from "ui";
import { cn } from "lib";
import { useWorkspacePath } from "@/hooks";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useStatuses } from "@/lib/hooks/statuses";
import { StoryStatusIcon } from "@/components/ui/story-status-icon";
import { getStoryPath } from "@/shared/routing/story";
import type { StoryChecklistItem } from "../types/checklist";

const ConvertedStoryLink = ({
  story,
}: {
  story: NonNullable<StoryChecklistItem["convertedStory"]>;
}) => {
  const { withWorkspace } = useWorkspacePath();
  const { data: teams = [] } = useTeams();
  const { data: statuses = [] } = useStatuses();
  const team = teams.find(({ id }) => id === story.teamId);
  const status = statuses.find(({ id }) => id === story.statusId);
  const reference = team?.code
    ? `${team.code}-${story.sequenceId}`
    : `#${story.sequenceId}`;

  return (
    <Tooltip
      title={
        <span className="block space-y-1 py-1">
          <span className="text-text-muted block">
            {reference}
            {status ? ` · ${status.name}` : ""}
          </span>
          <span className="block font-medium">{story.title}</span>
          {team ? (
            <span className="text-text-muted block">{team.name}</span>
          ) : null}
        </span>
      }
    >
      <Link
        aria-label={`${reference}: ${story.title}`}
        className="focus-visible:ring-ring inline-flex shrink-0 rounded focus-visible:ring-2 focus-visible:outline-none"
        href={withWorkspace(getStoryPath({ ...story, teamCode: team?.code }))}
      >
        <Badge
          className="inline-flex items-center gap-1.5 px-2"
          color="tertiary"
        >
          <StoryStatusIcon statusId={story.statusId ?? undefined} />
          {reference}
        </Badge>
      </Link>
    </Tooltip>
  );
};

export const ChecklistItemContent = ({
  item,
  isConverted,
  canEdit,
  isPending,
  storyTerm,
  onEdit,
}: {
  item: StoryChecklistItem;
  isConverted: boolean;
  canEdit: boolean;
  isPending: boolean;
  storyTerm: string;
  onEdit: () => void;
}) => {
  const convertedStory = item.convertedStory;
  return (
    <div className="flex min-h-9 min-w-0 flex-1 items-center gap-2">
      {isConverted ? (
        <div className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 py-1">
          <span className="wrap-anywhere">{item.title}</span>
          <span className="text-text-muted italic">was converted to</span>
          {convertedStory ? (
            <ConvertedStoryLink story={convertedStory} />
          ) : (
            <span className="text-text-muted">an unavailable {storyTerm}</span>
          )}
        </div>
      ) : null}
      {!isConverted && canEdit ? (
        <button
          aria-label={`Edit ${item.title}`}
          className={cn(
            "hover:bg-state-hover focus-visible:ring-ring -ml-1 min-w-0 rounded px-1 py-1 text-left wrap-anywhere focus-visible:ring-2 focus-visible:outline-none",
            { "text-text-muted line-through": item.completed },
          )}
          disabled={isPending}
          onClick={onEdit}
          type="button"
        >
          {item.title}
        </button>
      ) : null}
      {!isConverted && !canEdit ? (
        <span
          className={cn("py-1 wrap-anywhere", {
            "text-text-muted line-through": item.completed,
          })}
        >
          {item.title}
        </span>
      ) : null}
    </div>
  );
};
