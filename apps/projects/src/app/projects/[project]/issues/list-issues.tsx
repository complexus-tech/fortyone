"use client";
import { BreadCrumbs, Button, Flex } from "ui";
import { PreferencesIcon, ProjectsIcon, IssuesIcon } from "icons";
import type { Issue, IssueStatus } from "@/types/issue";
import { useLocalStorage } from "@/hooks";
import type { IssuesLayout } from "@/components/ui";
import { IssuesBoard, LayoutSwitcher, NewIssueButton } from "@/components/ui";
import { HeaderContainer } from "@/components/layout";

export const ListIssues = ({
  issues,
  statuses,
}: {
  issues: Issue[];
  statuses: IssueStatus[];
}) => {
  const [layout, setLayout] = useLocalStorage<IssuesLayout>(
    "project:issues:layout",
    "kanban",
  );

  return (
    <>
      <HeaderContainer className="justify-between">
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Web design",
              icon: <ProjectsIcon className="h-4 w-auto" />,
            },
            {
              name: "Issues",
              icon: <IssuesIcon className="h-[1.1rem] w-auto" />,
            },
          ]}
        />
        <Flex gap={2}>
          <LayoutSwitcher layout={layout} setLayout={setLayout} />
          <Button
            color="tertiary"
            leftIcon={<PreferencesIcon className="h-4 w-auto" />}
            size="sm"
            variant="outline"
          >
            Display
          </Button>
          <NewIssueButton />
        </Flex>
      </HeaderContainer>
      <IssuesBoard issues={issues} layout={layout} statuses={statuses} />
    </>
  );
};
