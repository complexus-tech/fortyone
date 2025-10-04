import React, { useMemo } from "react";
import { Header } from "./components";
import { SafeContainer, StoriesListSkeleton } from "@/components/ui";
import { StoriesBoard } from "@/modules/stories/components";
import { useSprintStoriesGrouped } from "@/modules/stories/hooks";
import { useSprintViewOptions } from "./hooks";
import { useGlobalSearchParams } from "expo-router";
import { useTerminology } from "@/hooks/use-terminology";

export const SprintStories = () => {
  const { sprintId, teamId } = useGlobalSearchParams<{
    sprintId: string;
    teamId: string;
  }>();
  const { viewOptions, isLoaded: viewOptionsLoaded } = useSprintViewOptions(
    sprintId!
  );
  const { getTermDisplay } = useTerminology();

  const queryOptions = useMemo(() => {
    return {
      groupBy: viewOptions.groupBy,
      orderBy: viewOptions.orderBy,
      orderDirection: viewOptions.orderDirection,
      teamIds: [teamId!],
    };
  }, [
    viewOptions.groupBy,
    viewOptions.orderBy,
    viewOptions.orderDirection,
    teamId,
  ]);

  const { data: groupedStories, isPending } = useSprintStoriesGrouped(
    sprintId!,
    viewOptions.groupBy,
    queryOptions
  );

  if (!viewOptionsLoaded) {
    return (
      <SafeContainer isFull>
        <Header />
        <StoriesListSkeleton />
      </SafeContainer>
    );
  }

  return (
    <SafeContainer isFull>
      <Header />
      <StoriesBoard
        groupedStories={groupedStories}
        groupFilters={queryOptions}
        isLoading={isPending}
        emptyTitle={`No ${getTermDisplay("storyTerm", { variant: "plural" })} found in this sprint`}
        emptyMessage={`There are no ${getTermDisplay("storyTerm", { variant: "plural" })} in this sprint at the moment.`}
      />
    </SafeContainer>
  );
};
