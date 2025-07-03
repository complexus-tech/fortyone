"use client";
import { useParams } from "next/navigation";
import { StoriesBoard, type StoriesLayout } from "@/components/ui";
import { useTeamStoriesGrouped } from "@/modules/stories/hooks/use-team-stories-grouped";
import { StoriesSkeleton } from "@/modules/teams/stories/stories-skeleton";
import { useTeamOptions } from "./provider";

export const AllStories = ({ layout }: { layout: StoriesLayout }) => {
  const { teamId } = useParams<{ teamId: string }>();
  const { viewOptions, filters } = useTeamOptions();
  const { data: groupedStories, isPending } = useTeamStoriesGrouped(
    teamId,
    viewOptions.groupBy,
    {
      priorities: filters.priorities ?? undefined,
      assigneeIds: filters.assigneeIds ?? undefined,
      categories: ["backlog"],
    },
  );

  if (isPending) {
    return <StoriesSkeleton layout={layout} />;
  }

  return (
    <StoriesBoard
      className="h-[calc(100dvh-4rem)]"
      groupedStories={groupedStories}
      layout={layout}
      viewOptions={viewOptions}
    />
  );
};
