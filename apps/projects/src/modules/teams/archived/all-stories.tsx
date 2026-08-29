"use client";
import { useParams } from "next/navigation";
import { StoriesBoard, type StoriesLayout } from "@/components/ui";
import { getGroupedStoryFilterParams } from "@/components/ui/stories-filter-query";
import { useTeamStoriesGrouped } from "@/modules/stories/hooks/use-team-stories-grouped";
import { StoriesSkeleton } from "@/modules/teams/stories/stories-skeleton";
import { useTeamOptions } from "./provider";

export const AllStories = ({ layout }: { layout: StoriesLayout }) => {
  const { teamId } = useParams<{ teamId: string }>();
  const { viewOptions, setViewOptions, filters } = useTeamOptions();
  const { data: groupedStories, isPending } = useTeamStoriesGrouped(
    teamId,
    viewOptions.groupBy,
    {
      orderBy: viewOptions.orderBy,
      orderDirection: viewOptions.orderDirection,
      ...getGroupedStoryFilterParams(filters),
      showSubStories: viewOptions.showSubStories ? true : undefined,
      teamIds: [teamId],
      includeArchived: true,
    },
  );

  if (isPending) {
    return <StoriesSkeleton className="h-full" layout={layout} />;
  }

  return (
    <StoriesBoard
      className="h-full"
      groupedStories={groupedStories}
      layout={layout}
      setViewOptions={setViewOptions}
      viewOptions={viewOptions}
    />
  );
};
