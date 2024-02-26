"use client";
import { BreadCrumbs, Button, Flex, Tooltip } from "ui";
import {
  PreferencesIcon,
  IssuesIcon,
  SprintsIcon,
  SidebarCollapseIcon,
  SidebarExpandIcon,
} from "icons";
import type { Issue, IssueStatus } from "@/types/issue";
import { useLocalStorage } from "@/hooks";
import type { IssuesLayout } from "@/components/ui";
import { IssuesBoard, LayoutSwitcher } from "@/components/ui";
import { HeaderContainer } from "@/components/layout";

export const ListIssues = ({
  issues,
  statuses,
}: {
  issues: Issue[];
  statuses: IssueStatus[];
}) => {
  const [layout, setLayout] = useLocalStorage<IssuesLayout>(
    "project:sprints:layout",
    "kanban",
  );
  const [isExpanded, setIsExpanded] = useLocalStorage<boolean>(
    "project:sprints:layout",
    true,
  );

  return (
    <>
      <HeaderContainer className="justify-between">
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Web design",
              icon: "🚀",
            },
            {
              name: "Sprints",
              icon: <SprintsIcon className="h-4 w-auto" />,
            },
            {
              name: "Issues",
              icon: <IssuesIcon className="h-[1.1rem] w-auto" />,
            },
          ]}
        />
        <Flex align="center" gap={2}>
          <LayoutSwitcher layout={layout} setLayout={setLayout} />
          <Button
            color="tertiary"
            leftIcon={<PreferencesIcon className="h-4 w-auto" />}
            size="sm"
            variant="outline"
          >
            Display
          </Button>
          <span className="text-gray-200 dark:text-dark-100">|</span>
          <Tooltip
            title={isExpanded ? "Hide sprint details" : "Show sprint details"}
          >
            <Button
              color="tertiary"
              leftIcon={
                isExpanded ? (
                  <SidebarCollapseIcon className="h-5 w-auto" />
                ) : (
                  <SidebarExpandIcon className="h-5 w-auto" />
                )
              }
              onClick={() => {
                setIsExpanded(!isExpanded);
              }}
              size="sm"
              variant="naked"
            >
              <span className="sr-only">
                {isExpanded ? "Hide sprint details" : "Show sprint details"}
              </span>
            </Button>
          </Tooltip>
        </Flex>
      </HeaderContainer>
      <IssuesBoard issues={issues} layout={layout} statuses={statuses} />
    </>
  );
};
