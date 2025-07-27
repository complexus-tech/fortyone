"use client";
import { Box } from "ui";
import { useState } from "react";
import type { Story as StoryType } from "@/modules/stories/types";
import { StoryRow } from "./story/row";
import { StoryDialog } from "./story-dialog";

export const StoriesList = ({
  isInSearch,
  stories,
  rowClassName,
}: {
  isInSearch?: boolean;
  stories: StoryType[];
  rowClassName?: string;
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
          className={rowClassName}
          handleStoryClick={(storyId) => {
            setStoryId(storyId);
            setIsDialogOpen(true);
          }}
          isInSearch={isInSearch}
          key={`${story.id}-${story.title.slice(0, 10)}`}
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
