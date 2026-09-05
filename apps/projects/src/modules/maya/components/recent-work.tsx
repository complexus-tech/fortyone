"use client";

import type { ReactNode } from "react";
import { useState } from "react";
import Link from "next/link";
import dynamic from "next/dynamic";
import { Box, Button, Skeleton, Tabs, Text } from "ui";
import { useUserRole } from "@/hooks/role";
import { useTerminology } from "@/hooks/use-terminology-display";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { StoriesBoard } from "@/components/ui/stories-board";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { WorkEmptyIllustration } from "@/components/ui/illustrations/empty-state-illustrations";
import { useWorkTab } from "../hooks/use-work-tab";
import { useRecentWork } from "../hooks/use-recent-work";
import {
  MAYA_WORK_TABS,
  RECENT_WORK_LIMIT,
  type MayaWorkTab,
} from "../utils/recent-work";
import { WorkAttention } from "./work-attention";

const NewStoryDialog = dynamic(
  () =>
    import("@/components/ui/new-story-dialog").then(
      (module) => module.NewStoryDialog,
    ),
  { ssr: false },
);

const RECENT_TASK_VIEW_OPTIONS: StoriesViewOptions = {
  groupBy: "none",
  orderBy: "updated",
  orderDirection: "desc",
  showEmptyGroups: false,
  showSubStories: false,
  displayColumns: ["ID", "Status", "Priority", "Assignee"],
};

const EmptySection = ({
  illustration,
  title,
  children,
  action,
}: {
  action: ReactNode;
  illustration: ReactNode;
  title: string;
  children: ReactNode;
}) => (
  <Box className="border-border/80 flex flex-col items-center border-t-[0.5px] px-5 py-8 text-center">
    {illustration}
    <Text className="mt-3" fontSize="xl" fontWeight="medium">
      {title}
    </Text>
    <Text className="mt-2 max-w-sm text-base leading-7" color="muted">
      {children}
    </Text>
    <Box className="mt-4">{action}</Box>
  </Box>
);

const RecentRowsSkeleton = () => (
  <Box
    aria-label="Loading recent work"
    className="space-y-5 py-4"
    role="status"
  >
    {Array.from({ length: RECENT_WORK_LIMIT }, (_, row) => (
      <Box className="flex items-center gap-3" key={row}>
        <Skeleton className="size-9 shrink-0 rounded-xl" />
        <Skeleton className="h-4 w-2/3 rounded sm:w-1/2" />
        <Skeleton className="ml-auto hidden h-4 w-20 rounded sm:block" />
      </Box>
    ))}
  </Box>
);

export const RecentWork = () => {
  const { withWorkspace } = useWorkspacePath();
  const { getTermDisplay } = useTerminology();
  const [tab, setTab] = useWorkTab();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const { userRole } = useUserRole();
  const canCreate = userRole === "admin" || userRole === "member";
  const { stories, isPending, isError, retry } = useRecentWork(tab);
  const storyTerm = getTermDisplay("storyTerm", { variant: "plural" });

  const tabLabels = {
    all: `All ${storyTerm}`,
    assigned: "Assigned",
    created: "Created",
  };
  const emptyTitle =
    tab === "all" ? `No recent ${storyTerm}` : `No ${tab} ${storyTerm}`;
  const capitalizedStoryTerm = getTermDisplay("storyTerm", {
    variant: "plural",
    capitalize: true,
  });
  const emptyDescriptions = {
    all: `${capitalizedStoryTerm} you open or work on will appear here.`,
    assigned: `${capitalizedStoryTerm} assigned to you will appear here.`,
    created: `${capitalizedStoryTerm} you create will appear here.`,
  };

  return (
    <Box className="mt-5 space-y-4 pb-6 md:mt-6">
      <WorkAttention />
      {isError ? (
        <Box className="flex items-center justify-between gap-3" role="alert">
          <Text color="muted" fontSize="sm">
            Some work couldn&apos;t be loaded.
          </Text>
          <Button color="tertiary" onClick={retry} size="sm" variant="naked">
            Retry
          </Button>
        </Box>
      ) : null}

      <Tabs
        onValueChange={(value) => {
          if (MAYA_WORK_TABS.includes(value as MayaWorkTab))
            setTab(value as MayaWorkTab);
        }}
        value={tab}
      >
        <Box className="mb-3 flex items-center justify-between gap-3">
          <Tabs.List
            aria-label="Work views"
            className="hide-scrollbar mx-0 max-w-full flex-nowrap overflow-x-auto md:mx-0"
          >
            {MAYA_WORK_TABS.map((value) => (
              <Tabs.Tab key={value} value={value}>
                {tabLabels[value]}
              </Tabs.Tab>
            ))}
          </Tabs.List>
          <Link
            className="shrink-0 underline-offset-4 hover:underline focus-visible:underline"
            href={withWorkspace(`/my-work?tab=${tab}`)}
          >
            View more work
          </Link>
        </Box>
        <Tabs.Panel value={tab}>
          {isPending ? <RecentRowsSkeleton /> : null}
          {!isPending && stories.length > 0 ? (
            <Box className="border-border/80 border-t-[0.5px]">
              <StoriesBoard
                className="h-auto overflow-visible pb-0"
                embedded
                groupedStories={{
                  groups: [
                    {
                      key: "none",
                      stories,
                      loadedCount: stories.length,
                      totalCount: stories.length,
                      hasMore: false,
                      nextPage: 1,
                    },
                  ],
                  meta: {
                    totalGroups: 1,
                    groupBy: "none",
                    orderBy: "updated",
                    orderDirection: "desc",
                    filters: {},
                  },
                }}
                layout="list"
                rowClassName="px-0 md:px-0"
                viewOptions={RECENT_TASK_VIEW_OPTIONS}
              />
            </Box>
          ) : null}
          {!isPending && !isError && stories.length === 0 ? (
            <EmptySection
              action={
                canCreate && tab !== "assigned" ? (
                  <Button
                    color="tertiary"
                    onClick={() => {
                      setIsCreateOpen(true);
                    }}
                    size="sm"
                  >
                    Create {getTermDisplay("storyTerm")}
                  </Button>
                ) : (
                  <Link
                    className="underline-offset-4 hover:underline focus-visible:underline"
                    href={withWorkspace("/my-work?tab=all")}
                  >
                    Browse all {storyTerm}
                  </Link>
                )
              }
              illustration={<WorkEmptyIllustration className="w-36" />}
              title={emptyTitle}
            >
              {emptyDescriptions[tab]}
            </EmptySection>
          ) : null}
        </Tabs.Panel>
      </Tabs>
      {isCreateOpen ? (
        <NewStoryDialog
          isOpen={isCreateOpen}
          onCreated={() => {
            setTab("created");
          }}
          setIsOpen={setIsCreateOpen}
        />
      ) : null}
    </Box>
  );
};
