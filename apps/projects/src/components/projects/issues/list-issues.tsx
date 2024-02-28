"use client";
import { BreadCrumbs, Button, Flex } from "ui";
import { PreferencesIcon, IssueIcon, ArrowDownIcon } from "icons";
import type { Issue, IssueStatus } from "@/types/issue";
import { useLocalStorage } from "@/hooks";
import { HeaderContainer } from "@/components/layout";
import type { IssuesLayout } from "@/components/ui";
import {
  IssuesBoard,
  LayoutSwitcher,
  NewIssueButton,
  SideDetailsSwitch,
  BoardDividedPanel,
} from "@/components/ui";
import { Sidebar } from "./sidebar";

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
  const [isExpanded, setIsExpanded] = useLocalStorage(
    "project:issues:expanded",
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
              name: "Issues",
              icon: <IssueIcon className="h-[1.1rem] w-auto" strokeWidth={2} />,
            },
          ]}
        />
        <Flex align="center" gap={2}>
          <LayoutSwitcher layout={layout} setLayout={setLayout} />
          <Button
            color="tertiary"
            leftIcon={<PreferencesIcon className="h-4 w-auto" />}
            rightIcon={<ArrowDownIcon className="h-4 w-auto" />}
            size="sm"
            variant="outline"
          >
            <span className="sr-only">Display</span>
          </Button>
          <NewIssueButton />
          <span className="text-gray-200 dark:text-dark-100">|</span>
          <SideDetailsSwitch
            isExpanded={isExpanded}
            setIsExpanded={setIsExpanded}
          />
        </Flex>
      </HeaderContainer>

      <BoardDividedPanel autoSaveId="project:issues:divided-panel">
        <BoardDividedPanel.MainPanel>
          <IssuesBoard issues={issues} layout={layout} statuses={statuses} />
        </BoardDividedPanel.MainPanel>
        <BoardDividedPanel.SideBar isExpanded={isExpanded}>
          <Sidebar />
        </BoardDividedPanel.SideBar>
      </BoardDividedPanel>
    </>
  );
};
