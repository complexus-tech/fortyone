import { useObjective } from "../hooks/use-objectives";
import { QueryState } from "@/components/ui/query-state";
import React, { useMemo } from "react";

import { SafeContainer, StoriesListSkeleton } from "@/components/ui";
import { StoriesBoard } from "@/modules/stories/components";
import { useObjectiveStoriesGrouped } from "@/modules/stories/hooks";

import { useLocalSearchParams } from "expo-router";
import { useTerminology } from "@/hooks/use-terminology";
import { useViewOptions } from "@/hooks/use-view-options";
import { useQueryClient } from "@tanstack/react-query";
import { storyKeys } from "@/constants/keys";
import { Header } from "./components";

export const ObjectiveStories = () => {
  const queryClient = useQueryClient();
  const { objectiveId, teamId } = useLocalSearchParams<{
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
  const context = useObjective(objectiveId);

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
    error,
    refetch,
    isRefetching,
  } = useObjectiveStoriesGrouped(
    objectiveId!,
    viewOptions.groupBy,
    queryOptions,
  );

  if (context.isPending || context.error || !context.data) {
    return (
      <SafeContainer isFull>
        <Header
          viewOptions={viewOptions}
          setViewOptions={setViewOptions}
          resetViewOptions={resetViewOptions}
        />
        <QueryState
          loading={context.isPending}
          title={
            context.isPending
              ? `Loading ${getTermDisplay("objectiveTerm")}`
              : `${getTermDisplay("objectiveTerm", { capitalize: true })} unavailable`
          }
          message={
            context.error?.message ||
            (context.isPending
              ? undefined
              : "This item may have been removed or you may no longer have access.")
          }
          onRetry={
            context.isPending
              ? undefined
              : () => {
                  void context.refetch();
                }
          }
        />
      </SafeContainer>
    );
  }

  if (!viewOptionsLoaded) {
    return (
      <SafeContainer isFull>
        <Header
          objective={context.data}
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
        objective={context.data}
        viewOptions={viewOptions}
        setViewOptions={setViewOptions}
        resetViewOptions={resetViewOptions}
      />
      <StoriesBoard
        groupedStories={groupedStories}
        groupFilters={queryOptions}
        isLoading={isPending}
        error={error}
        onRetry={() => {
          void refetch();
        }}
        visibleColumns={viewOptions.displayColumns}
        emptyTitle={`No ${getTermDisplay("storyTerm", { variant: "plural" })} found for this ${getTermDisplay("objectiveTerm", { variant: "singular" })}`}
        emptyMessage={`There are no ${getTermDisplay("storyTerm", { variant: "plural" })} for this ${getTermDisplay("objectiveTerm", { variant: "singular" })} at the moment.`}
        onRefresh={async () => {
          await queryClient.invalidateQueries({ queryKey: storyKeys.all });
        }}
        isRefreshing={isRefetching}
      />
    </SafeContainer>
  );
};
