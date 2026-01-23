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
import { useTerminology } from "@/hooks";
import type { StateCategory } from "@/types/states";
import { useMyStoriesGrouped } from "@/modules/stories/hooks/use-my-stories-grouped";
import { MyWorkSkeleton } from "@/modules/my-work/components/my-work-skeleton";
import { useMyWork } from "./provider";

export const ListMyWork = ({ layout }: { layout: StoriesLayout }) => {
  const tabs = ["all", "assigned", "created"] as const;
  const [tab, setTab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("all"),
  );
  const validCategories = [
    "backlog",
    "unstarted",
    "started",
    "paused",
    "completed",
    "cancelled",
  ] as const satisfies ReadonlyArray<StateCategory>;
  const [category] = useQueryState(
    "category",
    parseAsStringLiteral(validCategories),
  );
  const [overdue] = useQueryState("overdue", parseAsBoolean);
  const [startDate] = useQueryState("startDate", parseAsIsoDate);
  const [endDate] = useQueryState("endDate", parseAsIsoDate);
  const { getTermDisplay } = useTerminology();
  const { viewOptions } = useMyWork();

  const categories: StateCategory[] | undefined = overdue
    ? ["started"]
    : category
      ? [category]
      : undefined;
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
      deadlineBefore: overdueDeadline,
      orderBy: viewOptions.orderBy,
    },
  );

  if (isPending) return <MyWorkSkeleton layout={layout} />;

  return (
    <Box className="h-[calc(100dvh-4rem)]">
      <Tabs onValueChange={(v) => setTab(v as typeof tab)} value={tab}>
        <Box className="border-border sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b-[0.5px]">
          <Tabs.List>
            <Tabs.Tab value="all">
              All {getTermDisplay("storyTerm", { variant: "plural" })}
            </Tabs.Tab>
            <Tabs.Tab value="assigned">Assigned</Tabs.Tab>
            <Tabs.Tab value="created">Created</Tabs.Tab>
          </Tabs.List>
        </Box>
        <Tabs.Panel value="all">
          <StoriesBoard
            className="h-[calc(100dvh-7.7rem)]"
            groupedStories={groupedStories}
            layout={layout}
            viewOptions={viewOptions}
          />
        </Tabs.Panel>
        <Tabs.Panel value="assigned">
          <StoriesBoard
            className="h-[calc(100dvh-7.7rem)]"
            groupedStories={groupedStories}
            layout={layout}
            viewOptions={viewOptions}
          />
        </Tabs.Panel>
        <Tabs.Panel value="created">
          <StoriesBoard
            className="h-[calc(100dvh-7.7rem)]"
            groupedStories={groupedStories}
            layout={layout}
            viewOptions={viewOptions}
          />
        </Tabs.Panel>
      </Tabs>
    </Box>
  );
};
