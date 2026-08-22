"use client";
import { addDays } from "date-fns";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { createContext, useContext, useEffect } from "react";
import type { ReactNode } from "react";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { getGroupedStoryFilterParams } from "@/components/ui/stories-filter-query";
import { useLocalStorage } from "@/hooks";
import type { StoriesLayout } from "@/components/ui";
import { useStoriesFilters } from "@/components/ui/stories-filter-state";
import type { StoriesFilter } from "@/components/ui/stories-filter-types";
import { useMyStoriesGrouped } from "@/modules/stories/hooks/use-my-stories-grouped";
import type { GroupedStoryParams } from "@/modules/stories/types";
import {
  ACTIVE_MY_WORK_CATEGORIES,
  getMyWorkDateValue,
  getMyWorkStoriesTotalCount,
  MY_WORK_TABS,
  STABLE_MY_WORK_TABS,
  type MyWorkTab,
} from "./tabs";

type MyWork = {
  viewOptions: StoriesViewOptions;
  setViewOptions: (value: StoriesViewOptions) => void;
  filters: StoriesFilter;
  setFilters: (value: StoriesFilter) => void;
  resetFilters: () => void;
  setTab: (value: MyWorkTab) => void;
  tab: MyWorkTab;
  visibleTabs: MyWorkTab[];
};

const MyWorkContext = createContext<MyWork | undefined>(undefined);

export const MyWorkProvider = ({
  children,
  layout,
}: {
  children: ReactNode;
  layout: StoriesLayout;
}) => {
  const initialOptions: StoriesViewOptions = {
    groupBy: "status",
    orderBy: "created",
    orderDirection: "desc",
    showEmptyGroups: true,
    showSubStories: true,
    displayColumns: [
      "ID",
      "Status",
      "Assignee",
      "Estimate",
      "Time needed",
      "Priority",
      "Deadline",
      "Created",
      "Updated",
      "Sprint",
      "Objective",
      "Key Result",
      "Labels",
    ],
  };
  const [viewOptions, setViewOptions] = useLocalStorage<StoriesViewOptions>(
    `my-work:view-options:${layout}`,
    initialOptions,
  );
  const { filters, resetFilters, setFilters } = useStoriesFilters();
  const [tab, setTabState] = useQueryState(
    "tab",
    parseAsStringLiteral(MY_WORK_TABS).withDefault("all"),
  );
  const countFilters = getGroupedStoryFilterParams(filters);
  const countOptions = {
    ...countFilters,
    assignedToMe: true,
    categories: [...ACTIVE_MY_WORK_CATEGORIES],
    showSubStories: viewOptions.showSubStories ? true : undefined,
    storiesPerGroup: 1,
  } satisfies Partial<GroupedStoryParams>;
  const today = getMyWorkDateValue(new Date());
  const tomorrow = getMyWorkDateValue(addDays(new Date(), 1));
  const nextWeek = getMyWorkDateValue(addDays(new Date(), 7));
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
  const todayCount = getMyWorkStoriesTotalCount(todayStories);
  const upcomingCount = getMyWorkStoriesTotalCount(upcomingStories);
  const blockedCount = getMyWorkStoriesTotalCount(blockedStories);
  const optionalTabs: MyWorkTab[] = [];

  if (todayCount > 0) optionalTabs.push("today");
  if (upcomingCount > 0) optionalTabs.push("upcoming");
  if (blockedCount > 0) optionalTabs.push("blocked");

  const visibleTabs = [
    "all",
    ...optionalTabs,
    ...STABLE_MY_WORK_TABS.slice(1),
  ] satisfies MyWorkTab[];
  const selectedTabIsVisible =
    STABLE_MY_WORK_TABS.includes(tab as (typeof STABLE_MY_WORK_TABS)[number]) ||
    (tab === "today" && (isTodayCountPending || todayCount > 0)) ||
    (tab === "upcoming" && (isUpcomingCountPending || upcomingCount > 0)) ||
    (tab === "blocked" && (isBlockedCountPending || blockedCount > 0));

  useEffect(() => {
    if (!selectedTabIsVisible) {
      void setTabState("all");
    }
  }, [selectedTabIsVisible, setTabState]);

  const setTab = (value: MyWorkTab) => {
    void setTabState(value);
  };

  return (
    <MyWorkContext.Provider
      value={{
        viewOptions,
        setViewOptions,
        filters,
        setFilters,
        resetFilters,
        setTab,
        tab,
        visibleTabs,
      }}
    >
      {children}
    </MyWorkContext.Provider>
  );
};

export const useMyWork = () => {
  const context = useContext(MyWorkContext);
  if (!context) {
    throw new Error("useMyWork must be used within a MyWorkProvider");
  }
  return context;
};
