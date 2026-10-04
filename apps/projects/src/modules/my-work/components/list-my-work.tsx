"use client";
import { Box, Button, Tabs, Text } from "ui";
import {
  parseAsBoolean,
  parseAsIsoDate,
  parseAsStringLiteral,
  useQueryState,
} from "nuqs";
import type { StoriesLayout } from "@/components/ui";
import { StoriesBoard } from "@/components/ui";
import { BoardSkeleton } from "@/components/ui/board-skeleton";
import { StoriesFilterBar } from "@/components/ui/stories-filter-bar";
import { StoriesEmptyIllustration } from "@/components/ui/illustrations/stories-empty-illustration";
import { useMyStoriesGrouped } from "@/modules/stories/hooks/use-my-stories-grouped";
import { walkthroughTargets } from "@/shared/walkthrough/targets";
import {
  getMyWorkDateValue,
  getMyWorkScopeFilterParams,
  MY_WORK_CATEGORIES,
} from "@/shared/story/my-work-scope";
import { getScopedStoriesFilterTeamId } from "@/components/ui/stories-filter-query";
import { getStoriesFilterOperator } from "@/components/ui/stories-filter-types";
import type { MyWorkViewScope } from "@/shared/story/my-work-scope";
import type { MyWorkTab } from "./tabs";
import { useMyWork } from "./provider";

const useMyWorkViewScope = (tab: MyWorkTab) => {
  const [category, setCategory] = useQueryState(
    "category",
    parseAsStringLiteral(MY_WORK_CATEGORIES),
  );
  const [overdue, setOverdue] = useQueryState("overdue", parseAsBoolean);
  const [startDate, setStartDate] = useQueryState("startDate", parseAsIsoDate);
  const [endDate, setEndDate] = useQueryState("endDate", parseAsIsoDate);
  const scope: MyWorkViewScope = {
    kind: "my-work",
    tab,
    category,
    overdue: overdue ?? false,
    createdAfter: startDate ? getMyWorkDateValue(startDate) : null,
    createdBefore: endDate ? getMyWorkDateValue(endDate) : null,
  };
  return {
    scope,
    clearScopeFilters: () => {
      void setCategory(null);
      void setOverdue(null);
      void setStartDate(null);
      void setEndDate(null);
    },
  };
};

const StoriesPanelContent = ({
  layout,
  tab,
}: {
  layout: StoriesLayout;
  tab: MyWorkTab;
}) => {
  const { scope } = useMyWorkViewScope(tab);
  const { viewOptions, setViewOptions, filters } = useMyWork();
  const { data: groupedStories, isPending } = useMyStoriesGrouped(
    viewOptions.groupBy,
    {
      ...getMyWorkScopeFilterParams(filters, scope),
      orderBy: viewOptions.orderBy,
      orderDirection: viewOptions.orderDirection,
      showSubStories: viewOptions.showSubStories ? true : undefined,
    },
  );
  return isPending ? (
    <BoardSkeleton className="h-full" layout={layout} />
  ) : (
    <StoriesBoard
      className="h-full"
      emptyStateIllustration={<StoriesEmptyIllustration />}
      groupedStories={groupedStories}
      layout={layout}
      setViewOptions={setViewOptions}
      viewOptions={viewOptions}
    />
  );
};

export const ListMyWork = ({ layout }: { layout: StoriesLayout }) => {
  const {
    filters,
    resetFilters,
    setFilters,
    tab,
    attentionStatus,
    retryAttention,
    viewOptions,
  } = useMyWork();

  const { scope, clearScopeFilters } = useMyWorkViewScope(tab);
  if (attentionStatus === "error") {
    return (
      <Box className="p-8 text-center">
        <Text role="alert">
          Unable to load the statuses needed for these filters.
        </Text>
        <Button onClick={retryAttention} size="sm" variant="outline">
          Retry
        </Button>
      </Box>
    );
  }
  if (attentionStatus === "pending") {
    return <BoardSkeleton layout={layout} />;
  }

  return (
    <Box
      className="h-(--app-page-content-height) min-h-0 overflow-hidden"
      data-walkthrough-target={walkthroughTargets.myWorkContent}
    >
      <Tabs className="flex h-full min-h-0 flex-col" value={tab}>
        <StoriesFilterBar
          filters={filters}
          resetFilters={() => {
            resetFilters();
            clearScopeFilters();
          }}
          saveView={{
            teamId: getScopedStoriesFilterTeamId(
              undefined,
              filters.teamIds,
              getStoriesFilterOperator(filters, "teamIds"),
            ),
            configuration: { version: 1, layout, filters, viewOptions, scope },
          }}
          setFilters={setFilters}
        />
        <Tabs.Panel className="min-h-0 flex-1" value="all">
          {tab === "all" ? (
            <StoriesPanelContent layout={layout} tab="all" />
          ) : null}
        </Tabs.Panel>
        <Tabs.Panel className="min-h-0 flex-1" value="today">
          {tab === "today" ? (
            <StoriesPanelContent layout={layout} tab="today" />
          ) : null}
        </Tabs.Panel>
        <Tabs.Panel className="min-h-0 flex-1" value="upcoming">
          {tab === "upcoming" ? (
            <StoriesPanelContent layout={layout} tab="upcoming" />
          ) : null}
        </Tabs.Panel>
        <Tabs.Panel className="min-h-0 flex-1" value="blocked">
          {tab === "blocked" ? (
            <StoriesPanelContent layout={layout} tab="blocked" />
          ) : null}
        </Tabs.Panel>
        <Tabs.Panel className="min-h-0 flex-1" value="assigned">
          {tab === "assigned" ? (
            <StoriesPanelContent layout={layout} tab="assigned" />
          ) : null}
        </Tabs.Panel>
        <Tabs.Panel className="min-h-0 flex-1" value="collaborating">
          {tab === "collaborating" ? (
            <StoriesPanelContent layout={layout} tab="collaborating" />
          ) : null}
        </Tabs.Panel>
        <Tabs.Panel className="min-h-0 flex-1" value="created">
          {tab === "created" ? (
            <StoriesPanelContent layout={layout} tab="created" />
          ) : null}
        </Tabs.Panel>
      </Tabs>
    </Box>
  );
};
