"use client";
import { useParams } from "next/navigation";
import { StoriesBoard, type StoriesLayout } from "@/components/ui";
import { useTeamStoriesGrouped } from "@/modules/stories/hooks/use-team-stories-grouped";
import { StoriesSkeleton } from "@/modules/teams/stories/stories-skeleton";
import { hasActiveStoriesFilters } from "@/components/ui/stories-filter-utils";
import { useTeamOptions } from "./provider";

export const AllStories = ({ layout }: { layout: StoriesLayout }) => {
  const { teamId } = useParams<{ teamId: string }>();
  const { viewOptions, filters } = useTeamOptions();
  const boardHeightClassName = hasActiveStoriesFilters(filters)
    ? "h-[calc(100dvh-7.2rem)]"
    : "h-[calc(100dvh-3.6rem)]";
  const { data: groupedStories, isPending } = useTeamStoriesGrouped(
    teamId,
    viewOptions.groupBy,
    {
      orderBy: viewOptions.orderBy,
      statusIds: filters.statusIds ?? undefined,
      priorities: filters.priorities ?? undefined,
      assigneeIds: filters.assigneeIds ?? undefined,
      reporterIds: filters.reporterIds ?? undefined,
      titleContains: filters.titleContains?.trim() || undefined,
      objectiveId: filters.objectiveId ?? undefined,
      startDateAfter: filters.startDate ?? undefined,
      startDateBefore: filters.startDate ?? undefined,
      deadlineAfter: filters.endDate ?? undefined,
      deadlineBefore: filters.endDate ?? undefined,
      hasNoAssignee: filters.hasNoAssignee ? true : undefined,
      showSubStories: viewOptions.showSubStories ? true : undefined,
      teamIds: [teamId],
      sprintIds: filters.sprintIds ?? undefined,
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
