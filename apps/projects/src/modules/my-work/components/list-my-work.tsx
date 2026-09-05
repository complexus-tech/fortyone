"use client";
import { Box, Button, Tabs, Text } from "ui";
import { formatISO } from "date-fns";
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
import { getGroupedStoryFilterParams } from "@/components/ui/stories-filter-query";
import type { StateCategory } from "@/types/states";
import { useMyStoriesGrouped } from "@/modules/stories/hooks/use-my-stories-grouped";
import {
  getStoryAttentionFilters,
  STORY_ATTENTION_VIEWS,
  type StoryAttentionView,
} from "@/shared/story/attention";
import { walkthroughTargets } from "@/shared/walkthrough/targets";
import { useMyWork } from "./provider";
import { getMyWorkTabFilterParams, type MyWorkTab } from "./tabs";

const StoriesPanelContent = ({
  layout,
  tab,
  attention,
}: {
  layout: StoriesLayout;
  tab: MyWorkTab;
  attention: StoryAttentionView | null;
}) => {
  const validCategories = [
    "backlog",
    "unstarted",
    "started",
    "paused",
    "completed",
    "cancelled",
  ] as const satisfies readonly StateCategory[];
  const [category] = useQueryState(
    "category",
    parseAsStringLiteral(validCategories),
  );
  const [overdue] = useQueryState("overdue", parseAsBoolean);
  const [startDate] = useQueryState("startDate", parseAsIsoDate);
  const [endDate] = useQueryState("endDate", parseAsIsoDate);
  const { viewOptions, setViewOptions, filters } = useMyWork();
  const tabFilters = getMyWorkTabFilterParams(tab, filters);
  const groupedFilters = getGroupedStoryFilterParams(filters);
  const hasEndDateFilter = Boolean(filters.endDate);

  let categories: StateCategory[] | undefined;
  if (overdue) {
    categories = ["started"];
  } else if (category) {
    categories = [category];
  }
  const overdueDeadline = overdue
    ? formatISO(new Date(), { representation: "date" })
    : undefined;
  const createdAfter = startDate
    ? formatISO(startDate, { representation: "date" })
    : undefined;
  const createdBefore = endDate
    ? formatISO(endDate, { representation: "date" })
    : undefined;

  const { data: groupedStories, isPending } = useMyStoriesGrouped(
    viewOptions.groupBy,
    {
      ...(attention
        ? getStoryAttentionFilters(attention, new Date())
        : {
            ...groupedFilters,
            ...tabFilters,
            categories: categories ?? tabFilters.categories,
            createdAfter: createdAfter ?? tabFilters.createdAfter,
            createdBefore: createdBefore ?? tabFilters.createdBefore,
            deadlineAfter: hasEndDateFilter
              ? groupedFilters.deadlineAfter
              : tabFilters.deadlineAfter,
            deadlineBefore: hasEndDateFilter
              ? groupedFilters.deadlineBefore
              : overdueDeadline ?? tabFilters.deadlineBefore,
            deadlineNot: hasEndDateFilter
              ? groupedFilters.deadlineNot
              : tabFilters.deadlineNot,
          }),
      orderBy: viewOptions.orderBy,
      orderDirection: viewOptions.orderDirection,
      showSubStories:
        attention || viewOptions.showSubStories ? true : undefined,
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
  const { filters, resetFilters, setFilters, tab } = useMyWork();
  const [attention, setAttention] = useQueryState(
    "attention",
    parseAsStringLiteral(STORY_ATTENTION_VIEWS),
  );

  return (
    <Box
      className="h-(--app-page-content-height) min-h-0 overflow-hidden"
      data-walkthrough-target={walkthroughTargets.myWorkContent}
    >
      <Tabs className="flex h-full min-h-0 flex-col" value={tab}>
        {attention ? (
          <Box className="border-border flex items-center justify-between gap-3 border-b px-5 py-2 md:px-12">
            <Text color="muted" fontSize="sm">
              Assigned to you ·{" "}
              {attention === "today" ? "Due today" : "Overdue"}
            </Text>
            <Button
              color="tertiary"
              onClick={() => {
                void setAttention(null);
              }}
              size="sm"
              variant="naked"
            >
              Clear
            </Button>
          </Box>
        ) : null}
        {!attention ? (
          <StoriesFilterBar
            filters={filters}
            resetFilters={resetFilters}
            setFilters={setFilters}
          />
        ) : null}
        <Tabs.Panel className="min-h-0 flex-1" value="all">
          {tab === "all" ? (
            <StoriesPanelContent
              attention={attention}
              layout={layout}
              tab="all"
            />
          ) : null}
        </Tabs.Panel>
        <Tabs.Panel className="min-h-0 flex-1" value="today">
          {tab === "today" ? (
            <StoriesPanelContent
              attention={attention}
              layout={layout}
              tab="today"
            />
          ) : null}
        </Tabs.Panel>
        <Tabs.Panel className="min-h-0 flex-1" value="upcoming">
          {tab === "upcoming" ? (
            <StoriesPanelContent
              attention={attention}
              layout={layout}
              tab="upcoming"
            />
          ) : null}
        </Tabs.Panel>
        <Tabs.Panel className="min-h-0 flex-1" value="blocked">
          {tab === "blocked" ? (
            <StoriesPanelContent
              attention={attention}
              layout={layout}
              tab="blocked"
            />
          ) : null}
        </Tabs.Panel>
        <Tabs.Panel className="min-h-0 flex-1" value="assigned">
          {tab === "assigned" ? (
            <StoriesPanelContent
              attention={attention}
              layout={layout}
              tab="assigned"
            />
          ) : null}
        </Tabs.Panel>
        <Tabs.Panel className="min-h-0 flex-1" value="collaborating">
          {tab === "collaborating" ? (
            <StoriesPanelContent
              attention={attention}
              layout={layout}
              tab="collaborating"
            />
          ) : null}
        </Tabs.Panel>
        <Tabs.Panel className="min-h-0 flex-1" value="created">
          {tab === "created" ? (
            <StoriesPanelContent
              attention={attention}
              layout={layout}
              tab="created"
            />
          ) : null}
        </Tabs.Panel>
      </Tabs>
    </Box>
  );
};
