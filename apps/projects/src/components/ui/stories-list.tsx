"use client";
import { Box } from "ui";
import { useState } from "react";
import type { Story as StoryType } from "@/modules/stories/types";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useBoardPropertySlots } from "@/shared/story/board-property-slots";
import { StoryRow } from "./story/row";
import { StoryDialog } from "./story-dialog";

export const StoriesList = ({
  isInSearch,
  stories,
  rowClassName,
  selectedCustomFieldIds,
}: {
  isInSearch?: boolean;
  stories: StoryType[];
  rowClassName?: string;
  selectedCustomFieldIds?: string[];
}) => {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [storyId, setStoryId] = useState<string | null>(null);
  const { Provider: BoardPropertyProvider } = useBoardPropertySlots();
  const { data: teams = [] } = useTeams();
  const teamCodesById = new Map(teams.map((team) => [team.id, team.code]));

  const handleNavigate = (newStoryId: string) => {
    setStoryId(newStoryId);
  };

  return (
    <BoardPropertyProvider
      selectedIds={selectedCustomFieldIds}
      stories={stories}
    >
      <Box>
        {stories.map((story) => (
          <StoryRow
            className={rowClassName}
            handleStoryClick={(storyId) => {
              setStoryId(storyId);
              setIsDialogOpen(true);
            }}
            isInSearch={isInSearch}
            key={`${story.id}-${story.title.slice(0, 10)}`}
            story={story}
            teamCode={story.team?.code ?? teamCodesById.get(story.teamId)}
          />
        ))}
        {storyId ? (
          <StoryDialog
            isOpen={isDialogOpen}
            onNavigate={handleNavigate}
            setIsOpen={setIsDialogOpen}
            stories={stories}
            storyId={storyId}
          />
        ) : null}
      </Box>
    </BoardPropertyProvider>
  );
};
