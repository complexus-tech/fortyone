"use client";
import { DndContext, DragOverlay } from "@dnd-kit/core";
import { Flex, Text } from "ui";
import { cn } from "lib";
import type { Issue, IssueStatus } from "@/types/issue";
import { BodyContainer } from "../layout/body";
import { KanbanBoard } from "./kanban-board";
import { IssuesGroup } from "./issues-group";
import { IssueStatusIcon } from "./issue-status-icon";
import { IssueCard } from "./issue/card";

export type IssuesLayout = "list" | "kanban" | null;

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
  return (
    <DndContext>
      {layout === "kanban" ? (
        <KanbanBoard
          className={className}
          issues={issues}
          statuses={statuses}
        />
      ) : (
        <BodyContainer className={cn("overflow-x-auto", className)}>
          {statuses.map((status) => (
            <IssuesGroup
              className="-top-[0.5px]"
              issues={issues}
              key={status}
              status={status}
            />
          ))}
        </BodyContainer>
      )}

      {layout === "kanban" ? (
        <DragOverlay
          dropAnimation={{
            duration: 300,
            easing: "cubic-bezier(0.18, 0.67, 0.6, 1.22)",
          }}
        >
          <IssueCard
            className="border-gray-200 shadow-lg dark:border-dark-50/60 dark:shadow-dark"
            issue={issues[0]}
          />
        </DragOverlay>
      ) : (
        <DragOverlay
          dropAnimation={{
            duration: 300,
            easing: "cubic-bezier(0.18, 0.67, 0.6, 1.22)",
          }}
        >
          <Flex
            align="center"
            className="w-max rounded-lg border border-gray-100 bg-gray-50/70 px-3 py-3.5 shadow backdrop-blur dark:border-dark-100 dark:bg-dark-200/70"
            gap={2}
          >
            <IssueStatusIcon status="Backlog" />
            <Text color="muted">COM-12</Text>
            <Text fontWeight="medium">Create a new project</Text>
          </Flex>
        </DragOverlay>
      )}
    </DndContext>
  );
};
