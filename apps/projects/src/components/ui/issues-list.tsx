"use client";
import { useDroppable } from "@dnd-kit/core";
import { cn } from "lib";
import type { Issue as IssueType } from "@/types/issue";
import { Issue } from "./issue/issue";

export const IssuesList = ({
  issues,
  id,
}: {
  issues: IssueType[];
  id: string | number;
}) => {
  const { isOver, setNodeRef } = useDroppable({
    id,
  });
  return (
    <div
      className={cn("border-0 border-transparent transition", {
        "border border-primary": isOver,
      })}
      ref={setNodeRef}
    >
      {issues.map((issue) => (
        <Issue issue={issue} key={issue.id} />
      ))}
    </div>
  );
};
