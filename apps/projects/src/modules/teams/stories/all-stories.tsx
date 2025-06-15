"use client";
import { Box, Tabs, Text } from "ui";
import { useParams } from "next/navigation";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { ArrowUpDownIcon, StoryIcon } from "icons";
import { StoriesBoard, StoryStatusIcon } from "@/components/ui";
import type { StoriesLayout } from "@/components/ui";
import { useTeamStoriesGrouped } from "@/modules/stories/hooks/use-team-stories-grouped";
import { StoriesSkeleton } from "@/modules/teams/stories/stories-skeleton";
import { useTeamOptions } from "./provider";

export const AllStories = ({ layout }: { layout: StoriesLayout }) => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: groupedStories, isPending } = useTeamStoriesGrouped(teamId);
  const { viewOptions } = useTeamOptions();

  const tabs = ["all", "active", "backlog"] as const;
  const [tab, setTab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("all"),
  );
  type Tab = (typeof tabs)[number];
  if (isPending) {
    return <StoriesSkeleton layout={layout} />;
  }

  return (
    <Tabs
      onValueChange={(v) => {
        setTab(v as Tab);
      }}
      value={tab}
    >
      <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full items-center justify-between border-b-[0.5px] border-gray-100 pr-12 dark:border-dark-100">
        <Tabs.List className="h-min">
          <Tabs.Tab leftIcon={<StoryIcon />} value="all">
            All stories
          </Tabs.Tab>
          <Tabs.Tab
            leftIcon={<StoryStatusIcon category="unstarted" />}
            value="active"
          >
            Active
          </Tabs.Tab>
          <Tabs.Tab leftIcon={<StoryStatusIcon />} value="backlog">
            Backlog
          </Tabs.Tab>
        </Tabs.List>
        <Text
          className="ml-2 hidden shrink-0 items-center gap-1.5 px-1 opacity-80 md:flex"
          color="muted"
        >
          <ArrowUpDownIcon className="h-4 w-auto" />
          Ordering by{" "}
          <span className="font-semibold capitalize">
            {viewOptions.orderBy}
          </span>
        </Text>
      </Box>

      <Tabs.Panel value="all">
        <StoriesBoard
          className="h-[calc(100dvh-7.7rem)]"
          groupedStories={groupedStories}
          layout={layout}
          stories={[]}
          viewOptions={viewOptions}
        />
      </Tabs.Panel>
      <Tabs.Panel value="active">
        <StoriesBoard
          className="h-[calc(100dvh-7.7rem)]"
          groupedStories={groupedStories}
          layout={layout}
          stories={[]}
          viewOptions={viewOptions}
        />
      </Tabs.Panel>
      <Tabs.Panel value="backlog">
        <StoriesBoard
          className="h-[calc(100dvh-7.7rem)]"
          groupedStories={groupedStories}
          layout={layout}
          stories={[]}
          viewOptions={viewOptions}
        />
      </Tabs.Panel>
    </Tabs>
  );
};
