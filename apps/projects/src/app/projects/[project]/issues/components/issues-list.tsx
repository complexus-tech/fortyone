"use client";
import type { IssueStatus, Issue as IssueType } from "@/types/issue";
import { BodyContainer } from "@/components/layout";
import { IssuesToolbar, IssuesGroup } from "@/components/ui";

export const IssuesList = ({
  issues,
  statuses,
}: {
  issues: IssueType[];
  statuses: IssueStatus[];
}) => {
  return (
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
  );
};
