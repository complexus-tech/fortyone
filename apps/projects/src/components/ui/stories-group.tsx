import { Box } from "ui";
import { cn } from "lib";
import type { StoryStatus, Story } from "@/types/story";
import { StoriesHeader } from "./stories-header";
import { StoriesList } from "./stories-list";

export const StoriesGroup = ({
  stories,
  status,
  className,
}: {
  stories: Story[];
  status: StoryStatus;
  className?: string;
}) => {
  const filteredStories = stories.filter((story) => story.status === status);

  return (
    <Box
      className={cn({
        hidden: filteredStories.length === 0,
      })}
    >
      <StoriesHeader
        className={className}
        count={filteredStories.length}
        status={status}
      />
      <StoriesList id={status} stories={filteredStories} />
    </Box>
  );
};
