import {
  Box,
  Flex,
  Button,
  Text,
  DatePicker,
  Tooltip,
  Badge,
  Divider,
} from "ui";
import {
  ArrowRightIcon,
  CalendarIcon,
  ObjectiveIcon,
  SprintsIcon,
  SubStoryIcon,
} from "icons";
import { cn } from "lib";
import { format, addDays, formatISO } from "date-fns";
import Link from "next/link";
import { ObjectivesMenu } from "@/components/ui/story/objectives-menu";
import { SprintsMenu } from "@/components/ui/story/sprints-menu";
import { Labels } from "@/components/ui/story/labels";
import { sprintTooltip } from "@/components/ui/story/sprint-tooltip";
import { getDueDateMessage } from "@/components/ui/story/due-date-tooltip";
import { StoryStatusIcon } from "@/components/ui/story-status-icon";
import { PriorityIcon } from "@/components/ui/priority-icon";
import { StatusesMenu } from "@/components/ui/story/statuses-menu";
import { PrioritiesMenu } from "@/components/ui/story/priorities-menu";
import type { Story } from "@/modules/stories/types";
import { useBoard } from "@/components/ui/board-context";
import type { StateCategory } from "@/types/states";
import { useTeamObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useMediaQuery, useTerminology, useUserRole } from "@/hooks";
import { useTeamSprints } from "@/modules/sprints/hooks/team-sprints";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { slugify } from "@/utils";
import { RowWrapper } from "../row-wrapper";

type StoryPropertiesProps = Story & {
  handleUpdate: (data: Partial<Story>) => void;
  asKanban?: boolean;
  teamCode?: string;
  isExpanded?: boolean;
  setIsExpanded?: (isExpanded: boolean) => void;
};

export const StoryProperties = ({
  handleUpdate,
  statusId,
  priority,
  objectiveId,
  sprintId,
  id,
  teamId,
  endDate,
  createdAt,
  updatedAt,
  labels: storyLabels,
  asKanban,
  subStories,
  teamCode,
  isExpanded,
  setIsExpanded,
}: StoryPropertiesProps) => {
  const { getTermDisplay } = useTerminology();
  const { isColumnVisible } = useBoard();
  const { data: statuses = [] } = useTeamStatuses(teamId);
  const { data: sprints = [] } = useTeamSprints(teamId);
  const { data: objectives = [] } = useTeamObjectives(teamId);
  const status =
    statuses.find((state) => state.id === statusId) || statuses.at(0);
  const selectedObjective = objectives.find(
    (objective) => objective.id === objectiveId,
  );
  const isMobile = useMediaQuery("(max-width: 768px)");
  const { userRole } = useUserRole();
  const isGuest = userRole === "guest";
  const selectedSprint = sprints.find((sprint) => sprint.id === sprintId);
  const completedOrCancelled = (category?: StateCategory) => {
    return ["completed", "cancelled", "paused"].includes(category || "");
  };
  return (
    <>
      {isColumnVisible("Status") && (
        <StatusesMenu>
          <StatusesMenu.Trigger>
            {asKanban ? (
              <Button
                className="gap-1 pr-2"
                color="tertiary"
                disabled={isGuest}
                size="xs"
                type="button"
                variant="outline"
              >
                <StoryStatusIcon className="h-[1.15rem]" statusId={statusId} />
                {status?.name}
              </Button>
            ) : (
              <button
                className="flex items-center gap-1 disabled:cursor-not-allowed disabled:opacity-50"
                disabled={isGuest}
                type="button"
              >
                <StoryStatusIcon statusId={statusId} />
                <span className="hidden md:inline">{status?.name}</span>
              </button>
            )}
          </StatusesMenu.Trigger>
          <StatusesMenu.Items
            setStatusId={(statusId) => {
              handleUpdate({ statusId });
            }}
            statusId={statusId}
            teamId={teamId}
          />
        </StatusesMenu>
      )}
      {isColumnVisible("Priority") && (
        <PrioritiesMenu>
          <PrioritiesMenu.Trigger>
            {asKanban ? (
              <Button
                className="gap-1 pr-2"
                color="tertiary"
                disabled={isGuest}
                size="xs"
                type="button"
                variant="outline"
              >
                <PriorityIcon className="h-[1.15rem]" priority={priority} />
                {priority}
              </Button>
            ) : (
              <button
                className="flex select-none items-center gap-1 disabled:cursor-not-allowed disabled:opacity-50"
                disabled={isGuest}
                type="button"
              >
                <PriorityIcon priority={priority} />
                <span className="hidden md:inline">{priority}</span>
              </button>
            )}
          </PrioritiesMenu.Trigger>
          <PrioritiesMenu.Items
            priority={priority}
            setPriority={(p) => {
              handleUpdate({ priority: p });
            }}
          />
        </PrioritiesMenu>
      )}
      {isColumnVisible("Objective") && selectedObjective ? (
        <ObjectivesMenu>
          <Tooltip
            className="max-w-80 py-3"
            title={
              <Flex align="start" gap={2}>
                <ObjectiveIcon className="relative top-[3px] h-4 shrink-0" />
                <Box>
                  <Text className="mb-1.5" fontSize="md">
                    {selectedObjective.name}
                  </Text>
                  <Text className="line-clamp-4" color="muted" fontSize="md">
                    {selectedObjective.description}
                  </Text>
                </Box>
              </Flex>
            }
          >
            <span>
              <ObjectivesMenu.Trigger>
                <Button
                  className={cn("gap-1 pr-2", {
                    "px-2": !asKanban,
                  })}
                  color="tertiary"
                  disabled={isGuest}
                  rounded={asKanban ? "md" : "xl"}
                  size="xs"
                  type="button"
                  variant="outline"
                >
                  <ObjectiveIcon className="h-4" />
                  <span className="inline-block max-w-32 truncate">
                    {selectedObjective.name}
                  </span>
                </Button>
              </ObjectivesMenu.Trigger>
            </span>
          </Tooltip>
          <ObjectivesMenu.Items
            objectiveId={objectiveId ?? undefined}
            setObjectiveId={(objectiveId) => {
              handleUpdate({ objectiveId });
            }}
            teamId={teamId}
          />
        </ObjectivesMenu>
      ) : null}
      {isColumnVisible("Sprint") && selectedSprint ? (
        <SprintsMenu>
          <Tooltip
            className="pointer-events-none max-w-96 py-3"
            title={sprintTooltip(selectedSprint)}
          >
            <span>
              <SprintsMenu.Trigger>
                <Button
                  className="gap-1 pr-2"
                  color="tertiary"
                  disabled={isGuest}
                  rounded={asKanban || isMobile ? "md" : "xl"}
                  size="xs"
                  type="button"
                  variant="outline"
                >
                  <SprintsIcon className="relative -top-[0.3px] h-[1.1rem]" />
                  <span className="inline-block max-w-36 truncate">
                    {selectedSprint.name}
                  </span>
                </Button>
              </SprintsMenu.Trigger>
            </span>
          </Tooltip>
          <SprintsMenu.Items
            setSprintId={(sprintId) => {
              handleUpdate({ sprintId });
            }}
            sprintId={sprintId ?? undefined}
            teamId={teamId}
          />
        </SprintsMenu>
      ) : null}
      {subStories.length > 0 && (
        <Tooltip
          title={
            <Box className="min-w-72">
              <Text
                className="mb-2.5 flex items-center gap-2 pt-1"
                fontSize="md"
              >
                <SubStoryIcon />
                Sub{" "}
                {getTermDisplay("storyTerm", {
                  variant: "plural",
                })}
              </Text>
              <Divider />
              {subStories.map((subStory) => (
                <Link
                  href={`/story/${subStory.id}/${slugify(subStory.title)}`}
                  key={subStory.id}
                >
                  <RowWrapper className="group max-w-72 gap-4 truncate px-0 py-2 last-of-type:border-b-0 dark:border-dark-50">
                    <Flex align="center" gap={2}>
                      <Text color="muted">
                        {teamCode}-{subStory.sequenceId}
                      </Text>
                      <Text className="group-hover:underline">
                        {subStory.title}
                      </Text>
                    </Flex>
                    <Flex align="center" gap={2}>
                      <PriorityIcon priority={subStory.priority} />
                      <StoryStatusIcon
                        className="h-[1.15rem]"
                        statusId={subStory.statusId}
                      />
                    </Flex>
                  </RowWrapper>
                </Link>
              ))}
            </Box>
          }
        >
          <Badge
            className="h-[1.85rem] cursor-pointer text-[0.95rem]"
            color="tertiary"
            onClick={() => {
              if (!asKanban) {
                setIsExpanded?.(!isExpanded);
              }
            }}
            role="button"
            rounded={asKanban ? "md" : "xl"}
            tabIndex={0}
          >
            <SubStoryIcon />
            {subStories.length <= 10 ? subStories.length : `10+`} sub{" "}
            {getTermDisplay("storyTerm", {
              variant: subStories.length === 1 ? "singular" : "plural",
            })}
            {!asKanban && (
              <ArrowRightIcon
                className={cn("h-3.5 transition-transform", {
                  "rotate-90": isExpanded,
                })}
              />
            )}
          </Badge>
        </Tooltip>
      )}
      {isColumnVisible("Labels") && storyLabels && storyLabels.length > 0 ? (
        <Labels
          isRectangular={asKanban}
          storyId={id}
          storyLabels={storyLabels}
          teamId={teamId}
        />
      ) : null}
      {isColumnVisible("Deadline") &&
      endDate &&
      !completedOrCancelled(status?.category) ? (
        <DatePicker>
          <Tooltip
            className="py-3"
            title={
              <Flex align="start" gap={2}>
                <CalendarIcon
                  className={cn("relative top-[2.5px] h-5 w-auto", {
                    "text-primary dark:text-primary":
                      new Date(endDate) < new Date(),
                    "text-warning dark:text-warning":
                      new Date(endDate) <= addDays(new Date(), 7) &&
                      new Date(endDate) >= new Date(),
                  })}
                />
                <Box>{getDueDateMessage(new Date(endDate))}</Box>
              </Flex>
            }
          >
            <span>
              <DatePicker.Trigger>
                <Button
                  className={cn("pr-2", {
                    "text-primary dark:text-primary":
                      new Date(endDate) < new Date(),
                    "text-warning dark:text-warning":
                      new Date(endDate) <= addDays(new Date(), 7) &&
                      new Date(endDate) >= new Date(),
                    "px-2": !asKanban,
                  })}
                  color="tertiary"
                  disabled={isGuest}
                  rounded={asKanban ? "md" : "xl"}
                  size="xs"
                  type="button"
                  variant="outline"
                >
                  <CalendarIcon
                    className={cn("h-4", {
                      "text-primary dark:text-primary":
                        new Date(endDate) < new Date(),
                      "text-warning dark:text-warning":
                        new Date(endDate) <= addDays(new Date(), 7) &&
                        new Date(endDate) >= new Date(),
                    })}
                    strokeWidth={3}
                  />
                  {format(new Date(endDate), "MMM d")}
                </Button>
              </DatePicker.Trigger>
            </span>
          </Tooltip>
          <DatePicker.Calendar
            onDayClick={(day) => {
              handleUpdate({ endDate: formatISO(day) });
            }}
            selected={new Date(endDate)}
          />
        </DatePicker>
      ) : null}
      {isColumnVisible("Created") && (
        <Tooltip
          title={`Created on ${format(new Date(createdAt), "MMM dd, yyyy HH:mm")}`}
        >
          <span className="cursor-default">
            <Text as="span" color="muted">
              {format(new Date(createdAt), "MMM dd")}
            </Text>
          </span>
        </Tooltip>
      )}
      {isColumnVisible("Updated") && (
        <Tooltip
          title={`Last updated on ${format(new Date(updatedAt), "MMM dd, yyyy HH:mm")}`}
        >
          <span className="cursor-default">
            <Text as="span" color="muted">
              {format(new Date(updatedAt), "MMM dd")}
            </Text>
          </span>
        </Tooltip>
      )}
    </>
  );
};
