"use client";
import { Box, Tabs } from "ui";
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
import { getGroupedStoryFilterParams } from "@/components/ui/stories-filter-query";
import { hasActiveStoriesFilters } from "@/components/ui/stories-filter-utils";
import { useTerminology, useUserRole } from "@/hooks";
import type { StateCategory } from "@/types/states";
import { useMyStoriesGrouped } from "@/modules/stories/hooks/use-my-stories-grouped";
import { PulseReportPanel } from "./pulse-report";
import { useMyWork } from "./provider";

const adminTabs = ["pulse", "all", "assigned", "created"] as const;
const storyTabs = ["all", "assigned", "created"] as const;

type MyWorkTab = (typeof adminTabs)[number];
type StoriesTab = Exclude<MyWorkTab, "pulse">;

const StoriesPanelContent = ({
  layout,
  tab,
}: {
  layout: StoriesLayout;
  tab: StoriesTab;
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
      createdByMe: tab === "created" || tab === "all" ? true : undefined,
      assignedToMe: tab === "assigned" || tab === "all" ? true : undefined,
      categories,
      createdAfter,
      createdBefore,
      ...getGroupedStoryFilterParams(filters),
      deadlineBefore: filters.endDate ?? overdueDeadline,
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
  const { userRole } = useUserRole();
  const isAdmin = userRole === "admin";
  const [tab, setTab] = useQueryState(
    "tab",
    parseAsStringLiteral(isAdmin ? adminTabs : storyTabs).withDefault(
      isAdmin ? "pulse" : "all",
    ),
  );
  const { getTermDisplay } = useTerminology();
  const { filters, resetFilters, setFilters } = useMyWork();

  return (
    <Box className="h-[calc(100dvh-4rem)]">
      <Tabs onValueChange={(v) => setTab(v as MyWorkTab)} value={tab}>
        <Box className="border-border sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b-[0.5px]">
          <Tabs.List>
            {isAdmin ? <Tabs.Tab value="pulse">Pulse</Tabs.Tab> : null}
            <Tabs.Tab value="all">
              All {getTermDisplay("storyTerm", { variant: "plural" })}
            </Tabs.Tab>
            <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
            <Tabs.Tab value="created">Created</Tabs.Tab>
          </Tabs.List>
        </Box>
        {tab !== "pulse" ? (
          <StoriesFilterBar
            filters={filters}
            resetFilters={resetFilters}
            setFilters={setFilters}
          />
        ) : null}
        {isAdmin ? (
          <Tabs.Panel value="pulse">
            {tab === "pulse" ? <PulseReportPanel /> : null}
          </Tabs.Panel>
        ) : null}
        <Tabs.Panel value="all">
          {tab === "all" ? (
            <StoriesPanelContent layout={layout} tab="all" />
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
