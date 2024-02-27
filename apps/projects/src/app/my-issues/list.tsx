"use client";
import { Box, Tabs } from "ui";
import type { IssueStatus, Issue } from "@/types/issue";
import { BodyContainer } from "@/components/layout";
import {
  IssuesGroup,
  IssuesToolbar,
  IssuesList,
  BoardDividedPanel,
} from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { Header } from "./header";

export const List = ({
  issues,
  statuses,
}: {
  issues: Issue[];
  statuses: IssueStatus[];
}) => {
  const [isExpanded, setIsExpanded] = useLocalStorage(
    "my-issues:expanded",
    true,
  );

  return (
    <>
      <Header isExpanded={isExpanded} setIsExpanded={setIsExpanded} />
      <BoardDividedPanel autoSaveId="my-issues:divided-panel">
        <BoardDividedPanel.MainPanel>
          <BodyContainer className="relative">
            <Tabs defaultValue="assigned">
              <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b border-gray-100 bg-white/70 backdrop-blur-lg dark:border-dark-200 dark:bg-dark-300/30">
                <Tabs.List>
                  <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
                  <Tabs.Tab value="created">Created</Tabs.Tab>
                  <Tabs.Tab value="subscribed">Subscribed</Tabs.Tab>
                </Tabs.List>
              </Box>
              <Tabs.Panel value="assigned">
                {statuses.map((status) => (
                  <IssuesGroup issues={issues} key={status} status={status} />
                ))}
              </Tabs.Panel>
              <Tabs.Panel value="created">
                <IssuesList issues={issues} />
              </Tabs.Panel>
              <Tabs.Panel value="subscribed">
                <IssuesList issues={issues} />
              </Tabs.Panel>
            </Tabs>
            <IssuesToolbar />
          </BodyContainer>
        </BoardDividedPanel.MainPanel>
        <BoardDividedPanel.SideBar isExpanded={isExpanded}>
          test 123
        </BoardDividedPanel.SideBar>
      </BoardDividedPanel>
    </>
  );
};
