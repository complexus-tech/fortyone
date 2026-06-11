"use client";
import { useParams } from "next/navigation";
import type { StoriesLayout } from "@/components/ui";
import { StoriesBoard } from "@/components/ui";
import { getGroupedStoryFilterParams } from "@/components/ui/stories-filter-query";
import { hasActiveStoriesFilters } from "@/components/ui/stories-filter-utils";
import { StoriesSkeleton } from "@/modules/teams/stories/stories-skeleton";
import { useSprintStoriesGrouped } from "@/modules/stories/hooks/use-sprint-stories-grouped";
import { useSprintOptions } from "./provider";

export const AllStories = ({ layout }: { layout: StoriesLayout }) => {
  const { sprintId, teamId } = useParams<{
    sprintId: string;
    teamId: string;
  }>();
  const { viewOptions, filters } = useSprintOptions();
  const boardHeightClassName = hasActiveStoriesFilters(filters)
    ? "h-[calc(100dvh-7.2rem)]"
    : "h-[calc(100dvh-3.6rem)]";
  const { data: groupedStories, isPending } = useSprintStoriesGrouped(
    sprintId,
    viewOptions.groupBy,
    {
      orderBy: viewOptions.orderBy,
      ...getGroupedStoryFilterParams(filters),
      teamIds: [teamId],
      sprintIds: [sprintId],
      showSubStories: viewOptions.showSubStories ? true : undefined,
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
      viewOptions={viewOptions}
    />
  );
};
