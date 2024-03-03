"use client";
import type { DragEndEvent, DragStartEvent } from "@dnd-kit/core";
import { DndContext, DragOverlay } from "@dnd-kit/core";
import { Flex, Text } from "ui";
import { cn } from "lib";
import { useState } from "react";
import { createPortal } from "react-dom";
import type { Issue, IssueStatus } from "@/types/issue";
import { BodyContainer } from "../shared/body";
import { KanbanBoard } from "./kanban-board";
import { IssuesGroup } from "./issues-group";
import { IssueStatusIcon } from "./issue-status-icon";
import { IssueCard } from "./issue/card";

export type IssuesLayout = "list" | "kanban" | null;

const IssueOverlay = ({
  issue,
  layout,
}: {
  issue: Issue | null;
  layout: IssuesLayout;
}) => {
  return (
    <DragOverlay
      dropAnimation={{
        duration: 300,
        easing: "cubic-bezier(0.18, 0.67, 0.6, 1.22)",
      }}
    >
      {layout === "kanban" ? (
        <IssueCard
          className="border-gray-200 shadow-lg dark:border-dark-50/60 dark:shadow-dark"
          issue={issue!}
        />
      ) : (
        <Flex
          align="center"
          className="w-max rounded-lg border border-gray-100 bg-gray-50/70 px-3 py-3.5 shadow backdrop-blur dark:border-dark-100 dark:bg-dark-200/70"
          gap={2}
        >
          <IssueStatusIcon status={issue?.status} />
          <Text color="muted">COM-{issue?.id}</Text>
          <Text className="max-w-xs truncate" fontWeight="medium">
            {issue?.title}
          </Text>
        </Flex>
      )}
    </DragOverlay>
  );
};

export const IssuesBoard = ({
  layout,
  issues,
  statuses,
  className,
}: {
  layout: IssuesLayout;
  issues: Issue[];
  statuses: IssueStatus[];
  className?: string;
}) => {
  const [issuesBoard, setIssuesBoard] = useState<Issue[]>(issues);
  const [activeIssue, setActiveIssue] = useState<Issue | null>(null);

  const handleDragStart = (e: DragStartEvent) => {
    const issue = issues.find(({ id }) => id === Number(e.active.id))!;
    setActiveIssue(issue);
  };

  const handleDragEnd = (e: DragEndEvent) => {
    const newStatus = e.over?.id as IssueStatus | null;
    if (newStatus) {
      const index = issuesBoard.findIndex(
        ({ id }) => id === Number(e.active.id),
      )!;
      issuesBoard[index].status = newStatus;
      setIssuesBoard([...issuesBoard]);
    }
    setActiveIssue(null);
  };

  return (
    <DndContext onDragEnd={handleDragEnd} onDragStart={handleDragStart}>
      {layout === "kanban" ? (
        <KanbanBoard
          className={className}
          issues={issuesBoard}
          statuses={statuses}
        />
      ) : (
        <BodyContainer className={cn("overflow-x-auto pb-6", className)}>
          {statuses.map((status) => (
            <IssuesGroup
              className="-top-[0.5px]"
              issues={issuesBoard}
              key={status}
              status={status}
            />
          ))}
        </BodyContainer>
      )}

      {typeof window !== "undefined" &&
        createPortal(
          <IssueOverlay issue={activeIssue} layout={layout} />,
          document.body,
        )}
    </DndContext>
  );
};
