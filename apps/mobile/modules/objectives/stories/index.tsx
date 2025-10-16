import React, { useMemo } from "react";

import { SafeContainer, StoriesListSkeleton } from "@/components/ui";
import { StoriesBoard } from "@/modules/stories/components";
import { useObjectiveStoriesGrouped } from "@/modules/stories/hooks";

import { useGlobalSearchParams } from "expo-router";
import { useTerminology } from "@/hooks/use-terminology";
import { useViewOptions } from "@/hooks/use-view-options";
import { useQueryClient } from "@tanstack/react-query";
import { storyKeys } from "@/constants/keys";
import { Header } from "./components";

export const ObjectiveStories = () => {
  const queryClient = useQueryClient();
  const { objectiveId, teamId } = useGlobalSearchParams<{
    objectiveId: string;
    teamId: string;
  }>();
  const {
    viewOptions,
    setViewOptions,
    resetViewOptions,
    isLoaded: viewOptionsLoaded,
  } = useViewOptions(`objective-${objectiveId}:view-options`);
  const { getTermDisplay } = useTerminology();

  const queryOptions = useMemo(() => {
    return {
      groupBy: viewOptions.groupBy,
      orderBy: viewOptions.orderBy,
      orderDirection: viewOptions.orderDirection,
      objectiveId: objectiveId!,
      teamIds: [teamId!],
    };
  }, [
    viewOptions.groupBy,
    viewOptions.orderBy,
    viewOptions.orderDirection,
    objectiveId,
    teamId,
  ]);

  const {
    data: groupedStories,
    isPending,
    refetch,
    isRefetching,
  } = useObjectiveStoriesGrouped(
    objectiveId!,
    viewOptions.groupBy,
    queryOptions
  );

  if (!viewOptionsLoaded) {
    return (
      <SafeContainer isFull>
        <Header
          viewOptions={viewOptions}
          setViewOptions={setViewOptions}
          resetViewOptions={resetViewOptions}
        />
        <StoriesListSkeleton />
      </SafeContainer>
    );
  }

  return (
    <SafeContainer isFull>
      <Header
        viewOptions={viewOptions}
        setViewOptions={setViewOptions}
        resetViewOptions={resetViewOptions}
      />
      <StoriesBoard
        groupedStories={groupedStories}
        groupFilters={queryOptions}
        isLoading={isPending}
        visibleColumns={viewOptions.displayColumns}
        emptyTitle={`No ${getTermDisplay("storyTerm", { variant: "plural" })} found for this ${getTermDisplay("objectiveTerm", { variant: "singular" })}`}
        emptyMessage={`There are no ${getTermDisplay("storyTerm", { variant: "plural" })} for this ${getTermDisplay("objectiveTerm", { variant: "singular" })} at the moment.`}
        onRefresh={() => {
          refetch();
          queryClient.invalidateQueries({ queryKey: storyKeys.all });
        }}
        isRefreshing={isRefetching}
      />
    </SafeContainer>
  );
};
