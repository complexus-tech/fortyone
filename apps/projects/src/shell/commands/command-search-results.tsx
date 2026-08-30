import type { CSSProperties, ReactNode } from "react";
import { EnterIcon, ObjectiveIcon, SearchIcon, StoryIcon } from "icons";
import { Box, Command, Flex, Kbd, Skeleton, Text } from "ui";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import { useStatuses } from "@/lib/hooks/statuses";
import type { Objective } from "@/modules/objectives/types";
import type { SearchResponse } from "@/modules/search/types";
import type { Story } from "@/modules/stories/types";
import { useTeams } from "@/modules/teams/hooks/teams";
import { hexToRgba } from "@/utils/color";

const RESULT_LIMIT = 5;
const SKELETON_TITLE_WIDTHS = ["w-3/5", "w-2/5", "w-1/2"] as const;
const RESULT_ICON_CLASSNAME =
  "flex size-8 shrink-0 items-center justify-center rounded-md border";
const RESULT_SHORTCUT_CLASSNAME = "ml-3 shrink-0";

const getStatusStyle = (color?: string): CSSProperties | undefined =>
  color
    ? {
        backgroundColor: hexToRgba(color, 0.1),
        borderColor: hexToRgba(color, 0.2),
        color,
      }
    : undefined;

const groupHeading = (label: string) => (
  <Text className="mb-1.5 pl-3 dark:antialiased" color="muted">
    {label}
  </Text>
);

const ResultItem = ({
  description,
  icon,
  iconClassName,
  iconStyle,
  label,
  onSelect,
  value,
}: {
  description: ReactNode;
  icon: React.ReactNode;
  iconClassName?: string;
  iconStyle?: CSSProperties;
  label: string;
  onSelect: () => void;
  value: string;
}) => (
  <Command.Item
    className="justify-between rounded-lg px-3 py-2.5 text-[1.05rem] opacity-90"
    onSelect={onSelect}
    value={value}
  >
    <Flex align="center" className="min-w-0" gap={3}>
      <span
        className={`${RESULT_ICON_CLASSNAME} ${iconClassName ?? "border-border bg-surface-muted text-icon"}`}
        style={iconStyle}
      >
        {icon}
      </span>
      <Box className="min-w-0">
        <Text className="truncate font-medium antialiased">{label}</Text>
        <Text className="truncate text-sm" color="muted">
          {description}
        </Text>
      </Box>
    </Flex>
    <Kbd className={RESULT_SHORTCUT_CLASSNAME}>
      <EnterIcon className="!size-3" />
    </Kbd>
  </Command.Item>
);

const WorkspaceSearchSkeleton = () => (
  <Box aria-label="Searching workspace" className="mb-3" role="status">
    <span className="sr-only">Searching workspace…</span>
    <Skeleton className="mb-2 ml-3 h-3 w-14 rounded" />
    {SKELETON_TITLE_WIDTHS.map((titleWidth) => (
      <Flex align="center" className="px-3 py-2.5" gap={3} key={titleWidth}>
        <Skeleton className="size-8 shrink-0 rounded-md" />
        <Box className="min-w-0 flex-1">
          <Skeleton className={`h-4 rounded ${titleWidth}`} />
          <Skeleton className="mt-2 h-3 w-24 rounded" />
        </Box>
        <Skeleton className="size-6 shrink-0 rounded-md" />
      </Flex>
    ))}
  </Box>
);

const StoryResults = ({
  onSelect,
  statusesById,
  stories,
  teamCodesById,
}: {
  onSelect: (story: Story) => void;
  statusesById: Map<string, { color: string; name: string }>;
  stories: Story[];
  teamCodesById: Map<string, string>;
}) => {
  if (stories.length === 0) return null;

  return (
    <Command.Group className="mb-3 px-0" heading={groupHeading("Tasks")}>
      {stories.slice(0, RESULT_LIMIT).map((story) => {
        const status = statusesById.get(story.statusId);
        const teamCode = story.team?.code ?? teamCodesById.get(story.teamId);
        const storyIdentifier = teamCode
          ? `${teamCode}-${story.sequenceId}`
          : `#${story.sequenceId}`;

        return (
          <ResultItem
            description={
              <>
                <span>{storyIdentifier}</span>
                {status ? (
                  <>
                    <span> · </span>
                    <span className="font-medium">{status.name}</span>
                  </>
                ) : null}
                <span> · </span>
                <span className="font-medium">{story.priority}</span>
              </>
            }
            icon={
              <StoryIcon className="h-4" style={{ color: status?.color }} />
            }
            iconStyle={getStatusStyle(status?.color)}
            key={story.id}
            label={story.title}
            onSelect={() => {
              onSelect(story);
            }}
            value={`story ${story.id}`}
          />
        );
      })}
    </Command.Group>
  );
};

const ObjectiveResults = ({
  objectives,
  onSelect,
  statusesById,
}: {
  objectives: Objective[];
  onSelect: (objective: Objective) => void;
  statusesById: Map<string, { color: string; name: string }>;
}) => {
  if (objectives.length === 0) return null;

  return (
    <Command.Group className="mb-3 px-0" heading={groupHeading("Objectives")}>
      {objectives.slice(0, RESULT_LIMIT).map((objective) => {
        const status = statusesById.get(objective.statusId);

        return (
          <ResultItem
            description={
              <>
                <span>Objective</span>
                {status ? (
                  <>
                    <span> · </span>
                    <span className="font-medium">{status.name}</span>
                  </>
                ) : null}
                {objective.health ? (
                  <>
                    <span> · </span>
                    <span>{objective.health}</span>
                  </>
                ) : null}
              </>
            }
            icon={
              <ObjectiveIcon className="h-4" style={{ color: status?.color }} />
            }
            iconStyle={getStatusStyle(status?.color)}
            key={objective.id}
            label={objective.name}
            onSelect={() => {
              onSelect(objective);
            }}
            value={`objective ${objective.id}`}
          />
        );
      })}
    </Command.Group>
  );
};

export const CommandSearchResults = ({
  hasSettledQuery,
  isError,
  isLoading,
  onSelectObjective,
  onSelectStory,
  onViewAll,
  query,
  results,
}: {
  hasSettledQuery: boolean;
  isError: boolean;
  isLoading: boolean;
  onSelectObjective: (objective: Objective) => void;
  onSelectStory: (story: Story) => void;
  onViewAll: () => void;
  query: string;
  results?: SearchResponse;
}) => {
  const { data: teams = [] } = useTeams();
  const { data: statuses = [] } = useStatuses();
  const { data: objectiveStatuses = [] } = useObjectiveStatuses();

  if (!query) return null;

  const teamCodesById = new Map(teams.map((team) => [team.id, team.code]));
  const statusesById = new Map(statuses.map((status) => [status.id, status]));
  const objectiveStatusesById = new Map(
    objectiveStatuses.map((status) => [status.id, status]),
  );

  const stories = results?.stories ?? [];
  const objectives = results?.objectives ?? [];
  const hasResults = stories.length > 0 || objectives.length > 0;

  return (
    <>
      {isLoading ? (
        <Command.Loading>
          <WorkspaceSearchSkeleton />
        </Command.Loading>
      ) : null}

      {isError ? (
        <Text className="px-3 py-4" color="muted">
          Workspace search is temporarily unavailable.
        </Text>
      ) : null}

      {hasSettledQuery && !isError ? (
        <>
          <StoryResults
            onSelect={onSelectStory}
            statusesById={statusesById}
            stories={stories}
            teamCodesById={teamCodesById}
          />
          <ObjectiveResults
            objectives={objectives}
            onSelect={onSelectObjective}
            statusesById={objectiveStatusesById}
          />
          {!hasResults ? (
            <Text className="px-3 py-3" color="muted">
              No matching tasks or objectives.
            </Text>
          ) : null}
        </>
      ) : null}

      <Command.Group className="mb-3 px-0">
        <Command.Item
          className="justify-between rounded-lg px-3 py-2.5 text-[1.05rem] opacity-90"
          onSelect={onViewAll}
          value={`search ${query}`}
        >
          <Flex align="center" className="min-w-0" gap={3}>
            <span
              className={`${RESULT_ICON_CLASSNAME} border-primary/20 bg-primary/10 text-primary`}
            >
              <SearchIcon className="h-4" />
            </span>
            <Text className="truncate font-medium antialiased">
              View all results for “{query}”
            </Text>
          </Flex>
          <Kbd className={RESULT_SHORTCUT_CLASSNAME}>
            <EnterIcon className="!size-3" />
          </Kbd>
        </Command.Item>
      </Command.Group>
    </>
  );
};
