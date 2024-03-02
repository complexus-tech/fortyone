import { Box, Flex, Tabs, Text } from "ui";
import { DndContext, DragOverlay } from "@dnd-kit/core";
import type { IssuesLayout } from "@/components/ui";
import { IssueStatusIcon, IssuesBoard } from "@/components/ui";
import type { IssueStatus, Issue } from "@/types/issue";

export const AllIssues = ({
  issues,
  statuses,
  layout,
}: {
  issues: Issue[];
  statuses: IssueStatus[];
  layout: IssuesLayout;
}) => {
  return (
    <Box className="h-[calc(100vh-4rem)]">
      <DndContext>
        <Tabs defaultValue="assigned">
          <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b border-gray-100 bg-white/70 backdrop-blur-lg dark:border-dark-200 dark:bg-dark-300/30">
            <Tabs.List>
              <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
              <Tabs.Tab value="created">Created</Tabs.Tab>
              <Tabs.Tab value="subscribed">Subscribed</Tabs.Tab>
            </Tabs.List>
          </Box>
          <Tabs.Panel value="assigned">
            <IssuesBoard
              className="h-[calc(100vh-7.7rem)]"
              issues={issues}
              layout={layout}
              statuses={statuses}
            />
          </Tabs.Panel>
          <Tabs.Panel value="created">
            <IssuesBoard
              className="h-[calc(100vh-7.7rem)]"
              issues={issues}
              layout={layout}
              statuses={statuses}
            />
          </Tabs.Panel>
          <Tabs.Panel value="subscribed">
            <IssuesBoard
              className="h-[calc(100vh-7.7rem)]"
              issues={issues}
              layout={layout}
              statuses={statuses}
            />
          </Tabs.Panel>
        </Tabs>

        <DragOverlay
          dropAnimation={{
            duration: 300,
            easing: "cubic-bezier(0.18, 0.67, 0.6, 1.22)",
          }}
        >
          <Flex
            align="center"
            className="w-max rounded-lg border border-gray-100 bg-gray-50/70 px-3 py-3.5 shadow backdrop-blur dark:border-dark-100 dark:bg-dark-200/70"
            gap={2}
          >
            <IssueStatusIcon status="Backlog" />
            <Text color="muted">COM-12</Text>
            <Text fontWeight="medium">Create a new project</Text>
          </Flex>
        </DragOverlay>
      </DndContext>
    </Box>
  );
};
