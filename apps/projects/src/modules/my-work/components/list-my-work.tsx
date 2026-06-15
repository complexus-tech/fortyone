"use client";
import { useEffect } from "react";
import { Box, Tabs } from "ui";
import { addDays, formatISO } from "date-fns";
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
import { getGroupedStoryFilterParams } from "@/components/ui/stories-filter-query";
import { hasActiveStoriesFilters } from "@/components/ui/stories-filter-utils";
import type { StoriesFilter } from "@/components/ui/stories-filter-types";
import { useTerminology } from "@/hooks";
import type { StateCategory } from "@/types/states";
import { useMyStoriesGrouped } from "@/modules/stories/hooks/use-my-stories-grouped";
import type {
  GroupedStoriesResponse,
  GroupedStoryParams,
} from "@/modules/stories/types";
import { useMyWork } from "./provider";

const tabs = [
  "all",
  "today",
  "upcoming",
  "blocked",
  "assigned",
  "created",
] as const;

type MyWorkTab = (typeof tabs)[number];

const stableTabs = ["all", "assigned", "created"] as const;

const activeCategories = [
  "backlog",
  "unstarted",
  "started",
  "paused",
] as const satisfies readonly StateCategory[];

const getStoriesTotalCount = (groupedStories?: GroupedStoriesResponse) =>
  groupedStories?.groups.reduce(
    (total, group) => total + group.totalCount,
    0,
  ) ?? 0;

const getDateValue = (date: Date) =>
  formatISO(date, { representation: "date" });

const getTabFilterParams = (
  tab: MyWorkTab,
  filters: StoriesFilter,
): Partial<GroupedStoryParams> => {
  const baseFilters = getGroupedStoryFilterParams(filters);
  const today = getDateValue(new Date());
  const tomorrow = getDateValue(addDays(new Date(), 1));
  const nextWeek = getDateValue(addDays(new Date(), 7));

  switch (tab) {
    case "today":
      return {
        ...baseFilters,
        assignedToMe: true,
        categories: [...activeCategories],
        deadlineAfter: today,
        deadlineBefore: today,
      };
    case "upcoming":
      return {
        ...baseFilters,
        assignedToMe: true,
        categories: [...activeCategories],
        deadlineAfter: tomorrow,
        deadlineBefore: nextWeek,
      };
    case "blocked":
      return {
        ...baseFilters,
        assignedToMe: true,
        categories: [...activeCategories],
        hasBlockedBy: true,
      };
    case "assigned":
      return {
        ...baseFilters,
        assignedToMe: true,
      };
    case "created":
      return {
        ...baseFilters,
        createdByMe: true,
      };
    case "all":
    default:
      return {
        ...baseFilters,
        createdByMe: true,
        assignedToMe: true,
      };
  }
};

const StoriesPanelContent = ({
  layout,
  tab,
}: {
  layout: StoriesLayout;
  tab: MyWorkTab;
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
  const tabFilters = getTabFilterParams(tab, filters);
  const groupedFilters = getGroupedStoryFilterParams(filters);

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
      ...groupedFilters,
      ...tabFilters,
      categories: categories ?? tabFilters.categories,
      createdAfter: createdAfter ?? tabFilters.createdAfter,
      createdBefore: createdBefore ?? tabFilters.createdBefore,
      deadlineBefore:
        filters.endDate ?? overdueDeadline ?? tabFilters.deadlineBefore,
      orderBy: viewOptions.orderBy,
      showSubStories: viewOptions.showSubStories ? true : undefined,
    },
  );
  const hasAppliedFilters = hasActiveStoriesFilters(filters);
  const boardHeightClassName = hasAppliedFilters
    ? "h-[calc(100dvh-11.3rem)]"
    : "h-[calc(100dvh-7.7rem)]";

  return isPending ? (
    <BoardSkeleton className={boardHeightClassName} layout={layout} />
  ) : (
    <StoriesBoard
      className={boardHeightClassName}
      groupedStories={groupedStories}
      layout={layout}
      setViewOptions={setViewOptions}
      viewOptions={viewOptions}
    />
  );
};

export const ListMyWork = ({ layout }: { layout: StoriesLayout }) => {
  const [tab, setTab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("all"),
  );
  const { getTermDisplay } = useTerminology();
  const { filters, resetFilters, setFilters, viewOptions } = useMyWork();
  const countFilters = getGroupedStoryFilterParams(filters);
  const countOptions = {
    ...countFilters,
    categories: [...activeCategories],
    assignedToMe: true,
    storiesPerGroup: 1,
    showSubStories: viewOptions.showSubStories ? true : undefined,
  } satisfies Partial<GroupedStoryParams>;
  const today = getDateValue(new Date());
  const tomorrow = getDateValue(addDays(new Date(), 1));
  const nextWeek = getDateValue(addDays(new Date(), 7));

  const { data: todayStories, isPending: isTodayCountPending } =
    useMyStoriesGrouped("none", {
      ...countOptions,
      deadlineAfter: today,
      deadlineBefore: today,
    });
  const { data: upcomingStories, isPending: isUpcomingCountPending } =
    useMyStoriesGrouped("none", {
      ...countOptions,
      deadlineAfter: tomorrow,
      deadlineBefore: nextWeek,
    });
  const { data: blockedStories, isPending: isBlockedCountPending } =
    useMyStoriesGrouped("none", {
      ...countOptions,
      hasBlockedBy: true,
    });
  const todayCount = getStoriesTotalCount(todayStories);
  const upcomingCount = getStoriesTotalCount(upcomingStories);
  const blockedCount = getStoriesTotalCount(blockedStories);
  const optionalTabs: MyWorkTab[] = [];
  if (todayCount > 0) {
    optionalTabs.push("today");
  }
  if (upcomingCount > 0) {
    optionalTabs.push("upcoming");
  }
  if (blockedCount > 0) {
    optionalTabs.push("blocked");
  }
  const visibleTabs = [
    "all",
    ...optionalTabs,
    ...stableTabs.slice(1),
  ] satisfies MyWorkTab[];
  const selectedTabIsVisible =
    stableTabs.includes(tab as (typeof stableTabs)[number]) ||
    (tab === "today" && (isTodayCountPending || todayCount > 0)) ||
    (tab === "upcoming" && (isUpcomingCountPending || upcomingCount > 0)) ||
    (tab === "blocked" && (isBlockedCountPending || blockedCount > 0));

  useEffect(() => {
    if (!selectedTabIsVisible) {
      void setTab("all");
    }
  }, [selectedTabIsVisible, setTab]);

  const tabLabels = {
    all: `All ${getTermDisplay("storyTerm", { variant: "plural" })}`,
    today: "Today",
    upcoming: "Upcoming",
    blocked: "Blocked",
    assigned: "Assigned",
    created: "Created",
  } satisfies Record<MyWorkTab, string>;

  return (
    <Box className="h-[calc(100dvh-4rem)]">
      <Tabs onValueChange={(v) => setTab(v as MyWorkTab)} value={tab}>
        <Box className="border-border sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b-[0.5px]">
          <Tabs.List>
            {visibleTabs.map((visibleTab) => (
              <Tabs.Tab key={visibleTab} value={visibleTab}>
                {tabLabels[visibleTab]}
              </Tabs.Tab>
            ))}
          </Tabs.List>
        </Box>
        <StoriesFilterBar
          filters={filters}
          resetFilters={resetFilters}
          setFilters={setFilters}
        />
        <Tabs.Panel value="all">
          {tab === "all" ? (
            <StoriesPanelContent layout={layout} tab="all" />
          ) : null}
        </Tabs.Panel>
        <Tabs.Panel value="today">
          {tab === "today" ? (
            <StoriesPanelContent layout={layout} tab="today" />
          ) : null}
        </Tabs.Panel>
        <Tabs.Panel value="upcoming">
          {tab === "upcoming" ? (
            <StoriesPanelContent layout={layout} tab="upcoming" />
          ) : null}
        </Tabs.Panel>
        <Tabs.Panel value="blocked">
          {tab === "blocked" ? (
            <StoriesPanelContent layout={layout} tab="blocked" />
          ) : null}
        </Tabs.Panel>
        <Tabs.Panel value="assigned">
          {tab === "assigned" ? (
            <StoriesPanelContent layout={layout} tab="assigned" />
          ) : null}
        </Tabs.Panel>
        <Tabs.Panel value="created">
          {tab === "created" ? (
            <StoriesPanelContent layout={layout} tab="created" />
          ) : null}
        </Tabs.Panel>
      </Tabs>
    </Box>
  );
};
