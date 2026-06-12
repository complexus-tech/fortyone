"use client";
import { useState } from "react";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { Box, Button, Tabs, Text } from "ui";
import { ArrowUpDownIcon, CopyIcon, ObjectiveIcon, StoryIcon } from "icons";
import { toast } from "sonner";
import { useParams } from "next/navigation";
import type { StoriesLayout } from "@/components/ui";
import { StoriesBoard } from "@/components/ui";
import { StoriesFilterBar } from "@/components/ui/stories-filter-bar";
import { getGroupedStoryFilterParams } from "@/components/ui/stories-filter-query";
import { hasActiveStoriesFilters } from "@/components/ui/stories-filter-utils";
import { useObjectiveOptions } from "@/modules/objectives/stories/provider";
import { useCopyToClipboard, useTerminology } from "@/hooks";
import { useObjectiveStoriesGrouped } from "@/modules/stories/hooks/use-objective-stories-grouped";
import { useObjective } from "@/modules/objectives/hooks";
import { ObjectivePageSkeleton } from "@/modules/objectives/stories/objective-page-skeleton";
import { Header } from "@/modules/objectives/stories/header";
import { StoriesSkeleton } from "@/modules/objectives/stories/stories-skeleton";
import { Overview } from "./overview";

export const AllStories = ({
  objectiveId,
  layout,
  setLayout,
}: {
  objectiveId: string;
  layout: StoriesLayout;
  setLayout: (layout: StoriesLayout) => void;
}) => {
  const { teamId } = useParams<{ teamId: string }>();
  const [isCopied, setIsCopied] = useState(false);

  const [_, copyText] = useCopyToClipboard();
  const { getTermDisplay } = useTerminology();
  const tabs = ["overview", "stories"] as const;
  const [tab, setTab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("overview"),
  );
  type Tab = (typeof tabs)[number];

  const { viewOptions, setViewOptions, filters, resetFilters, setFilters } =
    useObjectiveOptions();
  const hasAppliedFilters = hasActiveStoriesFilters(filters);
  const boardHeightClassName = hasAppliedFilters
    ? "h-[calc(100dvh-11.3rem)]"
    : "h-[calc(100dvh-7.7rem)]";
  const { isPending: isObjectivePending } = useObjective(objectiveId);
  const { isPending: isStoriesPending, data: groupedStories } =
    useObjectiveStoriesGrouped(objectiveId, viewOptions.groupBy, {
      orderBy: viewOptions.orderBy,
      ...getGroupedStoryFilterParams(filters),
      teamIds: [teamId],
      objectiveId,
      showSubStories: viewOptions.showSubStories ? true : undefined,
    });

  if (isObjectivePending) {
    return <ObjectivePageSkeleton layout={layout} />;
  }

  return (
    <Box>
      <Header layout={layout} setLayout={setLayout} />
      <Tabs onValueChange={(v) => setTab(v as Tab)} value={tab}>
        <Box className="border-border sticky top-0 z-10 flex h-[3.7rem] w-full items-center justify-between border-b-[0.5px] pr-6 md:pr-12">
          <Tabs.List className="h-min">
            <Tabs.Tab leftIcon={<ObjectiveIcon />} value="overview">
              Overview
            </Tabs.Tab>
            <Tabs.Tab leftIcon={<StoryIcon />} value="stories">
              {getTermDisplay("storyTerm", {
                variant: "plural",
                capitalize: true,
              })}
            </Tabs.Tab>
          </Tabs.List>
          {tab !== "overview" ? (
            <Text
              className="ml-2 hidden shrink-0 items-center gap-1.5 px-1 opacity-80 md:flex"
              color="muted"
            >
              <ArrowUpDownIcon className="h-4 w-auto" />
              Ordering by <b className="capitalize">{viewOptions.orderBy}</b>
            </Text>
          ) : (
            <Button
              className="gap-1 px-3"
              color="tertiary"
              leftIcon={<CopyIcon className="h-4" />}
              onClick={async () => {
                await copyText(window.location.href);
                setIsCopied(true);
                toast.info("Success", {
                  description: `${getTermDisplay("objectiveTerm", { capitalize: true })} link copied to clipboard`,
                });
                setTimeout(() => {
                  setIsCopied(false);
                }, 5000);
              }}
              size="sm"
            >
              <span className="hidden md:inline">
                {isCopied ? "Copied" : "Copy link"}
              </span>
              <span className="md:hidden">{isCopied ? "Copied" : "Copy"}</span>
            </Button>
          )}
        </Box>
        <Tabs.Panel value="overview">
          <Overview />
        </Tabs.Panel>
        <Tabs.Panel value="stories">
          <StoriesFilterBar
            filters={filters}
            hiddenFields={["teamIds", "objectiveId"]}
            resetFilters={resetFilters}
            setFilters={setFilters}
          />
          {isStoriesPending ? (
            <StoriesSkeleton className={boardHeightClassName} layout={layout} />
          ) : (
            <StoriesBoard
              className={boardHeightClassName}
              groupedStories={groupedStories}
              layout={layout}
              setViewOptions={setViewOptions}
              viewOptions={viewOptions}
            />
          )}
        </Tabs.Panel>
      </Tabs>
    </Box>
  );
};
