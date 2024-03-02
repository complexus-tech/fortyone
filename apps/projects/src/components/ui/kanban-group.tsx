"use client";
import type { ReactNode } from "react";
import { useDroppable } from "@dnd-kit/core";
import { cn } from "lib";
import type { Issue, IssueStatus } from "@/types/issue";
import { IssueCard } from "./issue/card";

const List = ({
  children,
  id,
}: {
  children: ReactNode;
  id: string | number;
}) => {
  const { isOver, setNodeRef } = useDroppable({
    id,
  });
  return (
    <div
      className={cn(
        "flex h-full flex-col gap-3 overflow-y-auto rounded-lg pb-6 transition",
        {
          "bg-white opacity-60 dark:bg-dark-200/60": isOver,
        },
      )}
      ref={setNodeRef}
    >
      {children}
    </div>
  );
};

export const KanbanGroup = ({
  issues,
  status,
}: {
  issues: Issue[];
  status: IssueStatus;
}) => {
  const filteredIssues = issues.filter((issue) => issue.status === status);
  return (
    <List id={status} key={status}>
      {filteredIssues.map((issue) => (
        <IssueCard issue={issue} key={issue.id} />
      ))}
    </List>
  );
};
