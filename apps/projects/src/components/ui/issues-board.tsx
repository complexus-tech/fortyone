import type { Issue, IssueStatus } from "@/types/issue";
import { BodyContainer } from "../layout/body";
import { KanbanBoard } from "./kanban-board";
import { IssuesGroup } from "./issues-group";
import { IssuesToolbar } from "./issues-toolbar";

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
    <>
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
    </>
  );
};
