"use client";
import Link from "next/link";
import { DatePicker, Flex, Text, Tooltip, Avatar } from "ui";
import { useDraggable } from "@dnd-kit/core";
import { cn } from "lib";
import type { Story as StoryProps } from "@/types/story";
import { RowWrapper } from "../row-wrapper";
import { StoryStatusIcon } from "../story-status-icon";
import { PriorityIcon } from "../priority-icon";
import { AssigneesMenu } from "./assignees-menu";
import { TableCheckbox } from "./checkbox";
import { StoryContextMenu } from "./context-menu";
import { DragHandle } from "./drag-handle";
import { Labels } from "./labels";
import { PrioritiesMenu } from "./priorities-menu";
import { StatusesMenu } from "./statuses-menu";

export const StoryRow = ({ story }: { story: StoryProps }) => {
  const { id, title, status = "Backlog", priority = "No Priority" } = story;
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id,
  });
  return (
    <div ref={setNodeRef}>
      <StoryContextMenu>
        <RowWrapper
          className={cn("gap-4", {
            "bg-gray-50 opacity-70 dark:bg-dark-50/40 dark:opacity-50":
              isDragging,
          })}
        >
          <Flex align="center" className="relative shrink select-none" gap={2}>
            <DragHandle {...listeners} {...attributes} />
            <TableCheckbox />
            <Tooltip title="Story ID: COM-12">
              <Text
                className="min-w-[6ch] truncate text-[0.98rem]"
                color="muted"
              >
                COM-{id}
              </Text>
            </Tooltip>
            <StatusesMenu>
              <StatusesMenu.Trigger>
                <button className="block" type="button">
                  <StoryStatusIcon status={status} />
                </button>
              </StatusesMenu.Trigger>
              <StatusesMenu.Items status={status} />
            </StatusesMenu>
            <Link href="/projects/web/stories/test-123-story">
              <Text className="line-clamp-1 hover:opacity-90">{title}</Text>
            </Link>
          </Flex>
          <Flex align="center" className="shrink-0" gap={3}>
            <Labels />
            <PrioritiesMenu>
              <PrioritiesMenu.Trigger>
                <button
                  className="flex select-none items-center gap-1"
                  type="button"
                >
                  <PriorityIcon priority={priority} />
                  {priority}
                </button>
              </PrioritiesMenu.Trigger>
              <PrioritiesMenu.Items priority={priority} />
            </PrioritiesMenu>
            <DatePicker>
              <DatePicker.Trigger>
                <button type="button">
                  <Tooltip title="Created on Sep 27, 2021">
                    <Text as="span" color="muted">
                      Sep 27
                    </Text>
                  </Tooltip>
                </button>
              </DatePicker.Trigger>
              <DatePicker.Calendar />
            </DatePicker>
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
              <AssigneesMenu.Items />
            </AssigneesMenu>
          </Flex>
        </RowWrapper>
      </StoryContextMenu>
    </div>
  );
};
