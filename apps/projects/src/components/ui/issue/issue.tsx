"use client";
import Link from "next/link";
import { Flex, Text, Tooltip } from "ui";
import { AssigneesMenu } from "./assignees-menu";
import { IssueCheckbox } from "./checkbox";
import { IssueContextMenu } from "./context-menu";
import { DragHandle } from "./drag-handle";
import { Labels } from "./labels";
import { PrioritiesMenu } from "./priorities-menu";
import { StatusesMenu } from "./statuses-menu";
import { IssueWrapper } from "./wrapper";
import type { Issue as IssueProps } from "@/types/issue";

export const Issue = ({
  title,
  status = "Backlog",
  priority = "No Priority",
}: IssueProps) => {
  return (
    <IssueContextMenu>
      <IssueWrapper>
        <Flex align="center" className="relative select-none" gap={2}>
          <DragHandle />
          <IssueCheckbox />
          <PrioritiesMenu isSearchEnabled priority={priority} />
          <Tooltip title="Issue ID: COM-12">
            <Text className="w-[55px] truncate" color="muted">
              COM-12
            </Text>
          </Tooltip>
          <StatusesMenu isSearchEnabled status={status} />
          <Link href="/issue">
            <Text className="overflow-hidden text-ellipsis whitespace-nowrap hover:opacity-90">
              {title}
            </Text>
          </Link>
        </Flex>
        <Flex align="center" gap={3}>
          <Labels />
          <Tooltip title="Created on Sep 27, 2021">
            <Text color="muted">Sep 27</Text>
          </Tooltip>
          <AssigneesMenu isSearchEnabled />
        </Flex>
      </IssueWrapper>
    </IssueContextMenu>
  );
};
