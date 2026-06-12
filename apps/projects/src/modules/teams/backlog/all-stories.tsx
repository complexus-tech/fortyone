"use client";
import { useParams } from "next/navigation";
import { StoriesBoard, type StoriesLayout } from "@/components/ui";
import { getGroupedStoryFilterParams } from "@/components/ui/stories-filter-query";
import { hasActiveStoriesFilters } from "@/components/ui/stories-filter-utils";
import { useTeamStoriesGrouped } from "@/modules/stories/hooks/use-team-stories-grouped";
import { StoriesSkeleton } from "@/modules/teams/stories/stories-skeleton";
import { useTeamOptions } from "./provider";

export const AllStories = ({ layout }: { layout: StoriesLayout }) => {
  const { teamId } = useParams<{ teamId: string }>();
  const { viewOptions, setViewOptions, filters } = useTeamOptions();
  const boardHeightClassName = hasActiveStoriesFilters(filters)
    ? "h-[calc(100dvh-7.2rem)]"
    : "h-[calc(100dvh-3.6rem)]";
  const { data: groupedStories, isPending } = useTeamStoriesGrouped(
    teamId,
    viewOptions.groupBy,
    {
      orderBy: viewOptions.orderBy,
      ...getGroupedStoryFilterParams(filters),
      categories: ["backlog"],
      showSubStories: viewOptions.showSubStories ? true : undefined,
      teamIds: [teamId],
    },
  );

  if (isPending) {
    return <StoriesSkeleton className={boardHeightClassName} layout={layout} />;
  }

  return (
    <StoriesBoard
      className={boardHeightClassName}
      groupedStories={groupedStories}
      layout={layout}
      setViewOptions={setViewOptions}
      viewOptions={viewOptions}
    />
  );
};
