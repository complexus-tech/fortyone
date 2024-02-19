"use client";
import Link from "next/link";
import { DatePicker, Flex, Text, Tooltip, Avatar } from "ui";
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
  const { title, status = "Backlog", priority = "No Priority" } = issue;
  return (
    <IssueContextMenu>
      <RowWrapper>
        <Flex align="center" className="relative flex-1 select-none" gap={2}>
          <DragHandle />
          <TableCheckbox />
          <PrioritiesMenu>
            <PrioritiesMenu.Trigger>
              <button className="block" type="button">
                <PriorityIcon priority={priority} />
              </button>
            </PrioritiesMenu.Trigger>
            <PrioritiesMenu.Items priority={priority} />
          </PrioritiesMenu>
          <Tooltip title="Issue ID: COM-12">
            <Text className="w-[55px] truncate" color="muted">
              COM-12
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
          <Link className="flex-1" href="/projects/web/issues/test-123-issue">
            <Text className=" overflow-hidden text-ellipsis whitespace-nowrap hover:opacity-90">
              {title}
            </Text>
          </Link>
        </Flex>
        <Flex align="center" gap={3}>
          <Labels />
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
  );
};
