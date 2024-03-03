"use client";
import Link from "next/link";
import { DatePicker, Flex, Text, Tooltip, Avatar } from "ui";
import { useDraggable } from "@dnd-kit/core";
import { cn } from "lib";
import type { Issue as IssueProps } from "@/types/issue";
import { RowWrapper } from "../row-wrapper";
import { IssueStatusIcon } from "../issue-status-icon";
import { PriorityIcon } from "../priority-icon";
import { AssigneesMenu } from "./assignees-menu";
import { TableCheckbox } from "./checkbox";
import { IssueContextMenu } from "./context-menu";
import { DragHandle } from "./drag-handle";
import { Labels } from "./labels";
import { PrioritiesMenu } from "./priorities-menu";
import { StatusesMenu } from "./statuses-menu";

export const Issue = ({ issue }: { issue: IssueProps }) => {
  const { id, title, status = "Backlog", priority = "No Priority" } = issue;
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id,
  });
  return (
    <div ref={setNodeRef}>
      <IssueContextMenu>
        <RowWrapper
          className={cn("gap-4", {
            "bg-gray-50 opacity-70 dark:bg-dark-50/40 dark:opacity-50":
              isDragging,
          })}
        >
          <Flex align="center" className="relative shrink select-none" gap={2}>
            <DragHandle {...listeners} {...attributes} />
            <TableCheckbox />
            <Tooltip title="Issue ID: COM-12">
              <Text className="w-[8ch] truncate text-[0.96rem]" color="muted">
                COM-{id}
              </Text>
            </Tooltip>
            <StatusesMenu>
              <StatusesMenu.Trigger>
                <button className="block" type="button">
                  <IssueStatusIcon status={status} />
                </button>
              </StatusesMenu.Trigger>
              <StatusesMenu.Items status={status} />
            </StatusesMenu>
            <Link href="/projects/web/issues/test-123-issue">
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
      </IssueContextMenu>
    </div>
  );
};
