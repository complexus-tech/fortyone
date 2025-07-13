"use client";
import { useParams } from "next/navigation";
import type { StoriesLayout } from "@/components/ui";
import { StoriesBoard } from "@/components/ui";
import { StoriesSkeleton } from "@/modules/teams/stories/stories-skeleton";
import { useSprintStoriesGrouped } from "@/modules/stories/hooks/use-sprint-stories-grouped";
import { useSprintOptions } from "./provider";

export const AllStories = ({ layout }: { layout: StoriesLayout }) => {
  const { sprintId, teamId } = useParams<{
    sprintId: string;
    teamId: string;
  }>();
  const { viewOptions } = useSprintOptions();
  const { data: groupedStories, isPending } = useSprintStoriesGrouped(
    sprintId,
    viewOptions.groupBy,
    {
      orderBy: viewOptions.orderBy,
      teamIds: [teamId],
    },
  );

  if (isPending) {
    return <StoriesSkeleton layout={layout} />;
  }

  return (
    <StoriesBoard
      groupedStories={groupedStories}
      layout={layout}
      viewOptions={viewOptions}
    />
  );
};
