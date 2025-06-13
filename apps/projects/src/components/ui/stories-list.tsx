"use client";
import { Box } from "ui";
import { useState } from "react";
import type { Story as StoryType } from "@/modules/stories/types";
import { StoryRow } from "./story/row";
import { StoryDialog } from "./story-dialog";

export const StoriesList = ({
  isInSearch,
  stories,
}: {
  isInSearch?: boolean;
  stories: StoryType[];
}) => {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [storyId, setStoryId] = useState<string | null>(null);

  const handleNavigate = (newStoryId: string) => {
    setStoryId(newStoryId);
  };

  return (
    <Box>
      {stories.map((story) => (
        <StoryRow
          handleStoryClick={(storyId) => {
            setStoryId(storyId);
            setIsDialogOpen(true);
          }}
          isInSearch={isInSearch}
          key={story.id}
          story={story}
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
  );
};
