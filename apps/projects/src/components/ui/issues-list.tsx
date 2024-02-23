import type { Issue as IssueType } from "@/types/issue";
import { Issue } from "./issue/issue";

export const IssuesList = ({ issues }: { issues: IssueType[] }) => {
  return (
    <>
      {issues.map((issue) => (
        <Issue issue={issue} key={issue.id} />
      ))}
    </>
  );
};
