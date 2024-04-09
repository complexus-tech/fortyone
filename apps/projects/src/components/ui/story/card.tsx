"use client";
import Link from "next/link";
import { Box, Flex, Button, Text, Avatar, DatePicker, Checkbox } from "ui";
import { CalendarIcon, TagsIcon } from "icons";
import { useDraggable } from "@dnd-kit/core";
import { cn } from "lib";
import type { Story as StoryProps } from "@/types/story";
import { StoryStatusIcon } from "../story-status-icon";
import { PriorityIcon } from "../priority-icon";
import { StoryContextMenu } from "./context-menu";
import { AssigneesMenu } from "./assignees-menu";
import { StatusesMenu } from "./statuses-menu";
import { PrioritiesMenu } from "./priorities-menu";

export const StoryCard = ({
  story,
  className,
}: {
  story: StoryProps;
  className?: string;
}) => {
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id: story.id,
  });
  return (
    <div ref={setNodeRef} {...listeners} {...attributes}>
      <StoryContextMenu>
        <Box
          className={cn(
            "w-[340px] cursor-pointer select-none rounded-[0.45rem] border-[0.5px] border-gray-100 bg-white px-4 py-3 backdrop-blur transition duration-200 ease-linear hover:bg-white/50 dark:border-dark-100/90 dark:bg-dark-300/60 dark:hover:bg-dark-200/60",
            {
              "bg-gray-50 opacity-70 dark:bg-dark-50/40 dark:opacity-50":
                isDragging,
            },
            className,
          )}
        >
          <Flex className="mb-0.5" gap={2} justify="between">
            <Flex align="center" gap={1}>
              <Checkbox className="rounded-[0.35rem]" />
              <Link className="flex-1" href="/teams/web/stories/test-123-story">
                <Text
                  className="w-[12ch] truncate text-[0.95rem]"
                  color="muted"
                  fontWeight="medium"
                >
                  COM-{story.id}
                </Text>
              </Link>
            </Flex>

            <AssigneesMenu>
              <AssigneesMenu.Trigger>
                <button className="block" type="button">
                  <Avatar
                    name="Joseph Mukorivo"
                    size="xs"
                    src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
                  />
                </button>
              </AssigneesMenu.Trigger>
              <AssigneesMenu.Items />
            </AssigneesMenu>
          </Flex>
          <Link className="flex-1" href="/teams/web/stories/test-123-story">
            <Text className="mb-2 line-clamp-2">{story.title}</Text>
          </Link>
          <Flex gap={1} wrap>
            <StatusesMenu>
              <StatusesMenu.Trigger>
                <Button
                  className="bg-white dark:border-dark-100 dark:bg-dark-200/30"
                  color="tertiary"
                  leftIcon={
                    <StoryStatusIcon
                      className="relative left-0.5 h-4 w-auto"
                      status={story.status}
                    />
                  }
                  size="xs"
                  type="button"
                  variant="outline"
                >
                  <span className="sr-only">{story.status}</span>
                </Button>
              </StatusesMenu.Trigger>
              <StatusesMenu.Items status={story.status} />
            </StatusesMenu>
            <PrioritiesMenu>
              <PrioritiesMenu.Trigger>
                <Button
                  className="bg-white dark:border-dark-100 dark:bg-dark-200/30"
                  color="tertiary"
                  leftIcon={
                    <PriorityIcon
                      className="relative left-0.5 h-4 w-auto"
                      priority={story.priority}
                    />
                  }
                  size="xs"
                  type="button"
                  variant="outline"
                >
                  <span className="sr-only">{story.priority}</span>
                </Button>
              </PrioritiesMenu.Trigger>
              <PrioritiesMenu.Items priority={story.priority} />
            </PrioritiesMenu>
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
          </Flex>
        </Box>
      </StoryContextMenu>
    </div>
  );
};
