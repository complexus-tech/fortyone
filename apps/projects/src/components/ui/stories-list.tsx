import { Box } from "ui";
import type { Story as StoryType } from "@/types/story";
import { StoryRow } from "./story/row";

export const StoriesList = ({ stories }: { stories: StoryType[] }) => {
  return (
    <Box>
      {stories.map((story) => (
        <StoryRow key={story.id} story={story} />
      ))}
    </Box>
  );
};
