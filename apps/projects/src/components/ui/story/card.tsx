"use client";
import Link from "next/link";
import {
  Box,
  Flex,
  Button,
  Text,
  Avatar,
  DatePicker,
  Tooltip,
  Checkbox,
} from "ui";
import { CalendarIcon, TagsIcon } from "icons";
import { useDraggable } from "@dnd-kit/core";
import { cn } from "lib";
import type { Story as StoryProps } from "@/modules/stories/types";
import { StoryStatusIcon } from "../story-status-icon";
import { PriorityIcon } from "../priority-icon";
import { useBoard } from "../board-context";
import { StoryContextMenu } from "./context-menu";
import { AssigneesMenu } from "./assignees-menu";
import { StatusesMenu } from "./statuses-menu";
import { PrioritiesMenu } from "./priorities-menu";
import { slugify } from "@/utils";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useStatuses } from "@/lib/hooks/statuses";
import { useSprints } from "@/lib/hooks/sprints";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";

export const StoryCard = ({
  story,
  className,
}: {
  story: StoryProps;
  className?: string;
}) => {
  const { id, title, sequenceId, priority, statusId, teamId } = story;
  const { data: teams = [] } = useTeams();
  const { data: statuses = [] } = useStatuses();
  const { data: sprints = [] } = useSprints();
  const { data: objectives = [] } = useObjectives();
  const activeStatus =
    statuses.find((state) => state.id === statusId) || statuses.at(0);
  const { code: teamCode } = teams.find((team) => team.id === teamId)!!;
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id,
  });

  const { isColumnVisible, selectedStories, setSelectedStories } = useBoard();
  return (
    <div ref={setNodeRef} {...listeners} {...attributes}>
      <StoryContextMenu story={story}>
        <Box
          className={cn(
            "w-[340px] cursor-pointer select-none rounded-xl border border-gray-100/80 bg-white px-4 pb-4 pt-3 shadow shadow-gray-100 backdrop-blur transition duration-200 ease-linear hover:bg-white/50 dark:border-dark-100/40 dark:bg-dark-300 dark:shadow-none dark:hover:bg-dark-200/60",
            {
              "bg-gray-50 opacity-70 dark:bg-dark-50/40 dark:opacity-50":
                isDragging,
            },
            className,
          )}
        >
          <Flex className="mb-2" gap={2} justify="between">
            <Flex align="center" gap={2}>
              <Checkbox
                checked={selectedStories.includes(story.id)}
                onCheckedChange={(checked) => {
                  setSelectedStories(
                    checked
                      ? [...selectedStories, story.id]
                      : selectedStories.filter(
                          (storyId) => storyId !== story.id,
                        ),
                  );
                }}
                className="rounded-[0.35rem]"
              />
              {isColumnVisible("ID") && (
                <Link
                  className="flex-1"
                  href={`/story/${id}/${slugify(title)}`}
                >
                  <Text
                    className="w-[12ch] truncate text-[0.95rem] uppercase"
                    color="muted"
                    fontWeight="medium"
                  >
                    {teamCode}-{sequenceId}
                  </Text>
                </Link>
              )}
            </Flex>
            {isColumnVisible("Assignee") && (
              <Tooltip
                className="py-3"
                title={
                  <Flex align="center" gap={2}>
                    <Avatar
                      name="Joseph Mukorivo"
                      src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
                    />
                    <Box>
                      <Text fontWeight="medium">Joseph Mukorivo</Text>
                      <Text color="muted">@josemukorivo</Text>
                    </Box>
                  </Flex>
                }
              >
                <span>
                  <AssigneesMenu>
                    <AssigneesMenu.Trigger>
                      <button className="flex" type="button">
                        <Avatar
                          name="Joseph Mukorivo"
                          size="xs"
                          src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
                        />
                      </button>
                    </AssigneesMenu.Trigger>
                    <AssigneesMenu.Items
                      onAssigneeSelected={(assigneeId) => {}}
                    />
                  </AssigneesMenu>
                </span>
              </Tooltip>
            )}
          </Flex>
          <Link className="flex-1" href={`/story/${id}/${slugify(title)}`}>
            <Text className="mb-3 line-clamp-2">{title}</Text>
          </Link>
          <Flex gap={1} wrap>
            {isColumnVisible("Status") && (
              <StatusesMenu>
                <StatusesMenu.Trigger>
                  <Button
                    className="bg-white dark:border-dark-100 dark:bg-dark-200/30"
                    color="tertiary"
                    leftIcon={
                      <StoryStatusIcon
                        className="relative left-0.5 h-4 w-auto"
                        statusId={statusId}
                      />
                    }
                    size="xs"
                    type="button"
                    variant="outline"
                  >
                    <span className="sr-only">{activeStatus?.name}</span>
                  </Button>
                </StatusesMenu.Trigger>
                <StatusesMenu.Items
                  statusId={statusId}
                  setStatusId={(s) => {}}
                />
              </StatusesMenu>
            )}
            {isColumnVisible("Priority") && (
              <PrioritiesMenu>
                <PrioritiesMenu.Trigger>
                  <Button
                    className="bg-white dark:border-dark-100 dark:bg-dark-200/30"
                    color="tertiary"
                    leftIcon={
                      <PriorityIcon
                        className="relative left-0.5 h-4 w-auto"
                        priority={priority}
                      />
                    }
                    size="xs"
                    type="button"
                    variant="outline"
                  >
                    <span className="sr-only">{priority}</span>
                  </Button>
                </PrioritiesMenu.Trigger>
                <PrioritiesMenu.Items
                  priority={priority}
                  setPriority={(pr) => {}}
                />
              </PrioritiesMenu>
            )}
            {isColumnVisible("Due date") && (
              <DatePicker>
                <DatePicker.Trigger>
                  <Button
                    className="bg-white px-2 text-sm dark:border-dark-100 dark:bg-dark-200/30"
                    color="tertiary"
                    leftIcon={<CalendarIcon className="h-4 w-auto" />}
                    size="xs"
                    variant="outline"
                  >
                    Sep 21
                  </Button>
                </DatePicker.Trigger>
                <DatePicker.Calendar />
              </DatePicker>
            )}
            {/* {isColumnVisible("Labels") && (
              <Button
                className="bg-white dark:border-dark-100 dark:bg-dark-200/30"
                color="tertiary"
                leftIcon={<TagsIcon className="h-4 w-auto" />}
                size="xs"
                type="button"
                variant="outline"
              >
                3 labels
              </Button>
            )} */}
          </Flex>
        </Box>
      </StoryContextMenu>
    </div>
  );
};
