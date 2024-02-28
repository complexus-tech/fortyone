"use client";
import { DndContext, DragOverlay } from "@dnd-kit/core";
import { Flex, Text } from "ui";
import type { Issue, IssueStatus } from "@/types/issue";
import { BodyContainer } from "../layout/body";
import { KanbanBoard } from "./kanban-board";
import { IssuesGroup } from "./issues-group";
import { IssuesToolbar } from "./issues-toolbar";
import { IssueStatusIcon } from "./issue-status-icon";
import { IssueCard } from "./issue/card";

export type IssuesLayout = "list" | "kanban" | null;

export const IssuesBoard = ({
  layout,
  issues,
  statuses,
}: {
  layout: IssuesLayout;
  issues: Issue[];
  statuses: IssueStatus[];
}) => {
  return (
    <DndContext>
      {layout === "kanban" ? (
        <KanbanBoard issues={issues} statuses={statuses} />
      ) : (
        <BodyContainer className="overflow-x-auto">
          {statuses.map((status) => (
            <IssuesGroup
              className="-top-[1px]"
              issues={issues}
              key={status}
              status={status}
            />
          ))}
          <IssuesToolbar />
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
            className="w-max rounded-lg border border-gray-100 bg-gray-50/70 px-3 py-4 shadow backdrop-blur dark:border-dark-100 dark:bg-dark-200/70"
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
