"use client";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { Box, Tabs } from "ui";
import { useParams } from "next/navigation";
import type { StoriesLayout } from "@/components/ui";
import { StoriesBoard } from "@/components/ui";
import { StoriesFilterBar } from "@/components/ui/stories-filter-bar";
import { getGroupedStoryFilterParams } from "@/components/ui/stories-filter-query";
import { useObjectiveOptions } from "@/modules/objectives/stories/provider";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { useObjectiveStoriesGrouped } from "@/modules/stories/hooks/use-objective-stories-grouped";
import { useObjective } from "@/modules/objectives/hooks";
import { ObjectivePageSkeleton } from "@/modules/objectives/stories/objective-page-skeleton";
import { ResourceNotFoundState } from "@/components/ui/resource-not-found-state";
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
  const { getTermDisplay } = useTerminology();
  const { withWorkspace } = useWorkspacePath();
  const tabs = ["overview", "stories"] as const;
  const [tab, setTab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("overview"),
  );
  type Tab = (typeof tabs)[number];

  const { viewOptions, setViewOptions, filters, resetFilters, setFilters } =
    useObjectiveOptions();
  const { data: objective, isPending: isObjectivePending } =
    useObjective(objectiveId);
  const { isPending: isStoriesPending, data: groupedStories } =
    useObjectiveStoriesGrouped(objectiveId, viewOptions.groupBy, {
      orderBy: viewOptions.orderBy,
      orderDirection: viewOptions.orderDirection,
      ...getGroupedStoryFilterParams(filters),
      teamIds: [teamId],
      objectiveId,
      showSubStories: viewOptions.showSubStories ? true : undefined,
    });

  if (isObjectivePending) {
    return <ObjectivePageSkeleton layout={layout} />;
  }

  if (!objective) {
    return (
      <ResourceNotFoundState
        actionLabel={`Go to ${getTermDisplay("objectiveTerm", { variant: "plural" })}`}
        description={`This ${getTermDisplay("objectiveTerm")} might not exist or you might not have access to it.`}
        href={withWorkspace(`/teams/${teamId}/objectives`)}
        title={`${getTermDisplay("objectiveTerm", { capitalize: true })} not found`}
      />
    );
  }

  return (
    <Box className="flex h-full min-h-0 flex-col overflow-hidden">
      <Tabs
        className="flex min-h-0 flex-1 flex-col overflow-hidden"
        onValueChange={(v) => setTab(v as Tab)}
        value={tab}
      >
        <Header layout={layout} setLayout={setLayout} />
        <Tabs.Panel className="min-h-0 flex-1 overflow-hidden" value="overview">
          <Overview key={objectiveId} />
        </Tabs.Panel>
        <Tabs.Panel
          className="min-h-0 flex-1 overflow-hidden data-[state=active]:flex data-[state=active]:flex-col"
          value="stories"
        >
          <Box className="shrink-0">
            <StoriesFilterBar
              filters={filters}
              hiddenFields={["teamIds", "objectiveId"]}
              resetFilters={resetFilters}
              setFilters={setFilters}
            />
          </Box>
          <Box className="min-h-0 flex-1 overflow-hidden">
            {isStoriesPending ? (
              <StoriesSkeleton className="h-full" layout={layout} />
            ) : (
              <StoriesBoard
                className="h-full"
                groupedStories={groupedStories}
                layout={layout}
                setViewOptions={setViewOptions}
                viewOptions={viewOptions}
              />
            )}
          </Box>
        </Tabs.Panel>
      </Tabs>
    </Box>
  );
};
