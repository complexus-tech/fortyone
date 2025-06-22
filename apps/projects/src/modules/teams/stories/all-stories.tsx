"use client";

import { useParams } from "next/navigation";
import { useTeamStories } from "@/modules/stories/hooks/team-stories";
import { useTeamOptions } from "@/modules/teams/stories/provider";
import { StoriesBoard } from "@/components/ui";
import type { StoriesLayout } from "@/components/ui";

export const AllStories = ({ layout }: { layout: StoriesLayout }) => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: stories = [] } = useTeamStories(teamId);

  const { viewOptions } = useTeamOptions();

  return (
    <StoriesBoard
      className="h-[calc(100dvh-4rem)]"
      layout={layout}
      stories={stories}
      viewOptions={viewOptions}
    />
  );
};
