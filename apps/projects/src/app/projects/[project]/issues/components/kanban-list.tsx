import type { ReactNode } from "react";
import type { IssueStatus } from "@/types/issue";

export const KanbanList = ({
  children,
  status,
}: {
  children: ReactNode;
  status: IssueStatus;
}) => {
  return (
    <div className="flex h-[calc(100vh-7.5rem)] flex-col gap-3 overflow-y-auto pb-6">
      {children}
    </div>
  );
};
