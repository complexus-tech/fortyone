"use client";
import { Box, Tabs } from "ui";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { useParams } from "next/navigation";
import type { StoriesLayout } from "@/components/ui";
import { StoriesBoard } from "@/components/ui";
import { StoriesFilterBar } from "@/components/ui/stories-filter-bar";
import { getGroupedStoryFilterParams } from "@/components/ui/stories-filter-query";
import { useGroupedStories } from "@/modules/stories/hooks/use-grouped-stories";
import { PROFILE_HIDDEN_FILTER_FIELDS } from "./filter-fields";
import { useProfile } from "./provider";
import { Skeleton } from "./skeleton";

export const AllStories = ({ layout }: { layout: StoriesLayout }) => {
  const { userId } = useParams<{
    userId: string;
  }>();
  const tabs = ["assigned", "created"] as const;
  const [tab, setTab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("assigned"),
  );
  const { viewOptions, setViewOptions, filters, resetFilters, setFilters } =
    useProfile();
  const { data: groupedStories, isPending } = useGroupedStories({
    groupBy: viewOptions.groupBy,
    ...getGroupedStoryFilterParams(filters),
    assigneeIds: tab === "assigned" ? [userId] : undefined,
    reporterIds: tab === "created" ? [userId] : undefined,
    orderBy: viewOptions.orderBy,
    orderDirection: viewOptions.orderDirection,
    showSubStories: viewOptions.showSubStories ? true : undefined,
  });

  if (isPending) return <Skeleton layout={layout} />;

  return (
    <Box className="h-(--app-page-content-height) min-h-0">
      <Tabs
        className="flex h-full min-h-0 flex-col"
        onValueChange={(v) => setTab(v as typeof tab)}
        value={tab}
      >
        <Box className="border-border d/40 sticky top-0 z-10 flex h-[3.7rem] w-full shrink-0 flex-col justify-center border-b-[0.5px] backdrop-blur-lg">
          <Tabs.List>
            <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
            <Tabs.Tab value="created">Created</Tabs.Tab>
          </Tabs.List>
        </Box>
        <StoriesFilterBar
          filters={filters}
          hiddenFields={PROFILE_HIDDEN_FILTER_FIELDS}
          resetFilters={resetFilters}
          setFilters={setFilters}
        />
        <Tabs.Panel className="min-h-0 flex-1" value="assigned">
          <StoriesBoard
            className="h-full"
            groupedStories={groupedStories}
            layout={layout}
            setViewOptions={setViewOptions}
            viewOptions={viewOptions}
          />
        </Tabs.Panel>
        <Tabs.Panel className="min-h-0 flex-1" value="created">
          <StoriesBoard
            className="h-full"
            groupedStories={groupedStories}
            layout={layout}
            setViewOptions={setViewOptions}
            viewOptions={viewOptions}
          />
        </Tabs.Panel>
      </Tabs>
    </Box>
  );
};
