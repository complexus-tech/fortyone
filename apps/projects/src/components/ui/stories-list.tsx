import { Box } from "ui";
import type { Story as StoryType } from "@/modules/stories/types";
import { StoryRow } from "./story/row";

export const StoriesList = ({
  isInSearch,
  stories,
}: {
  isInSearch?: boolean;
  stories: StoryType[];
}) => {
  return (
    <Box>
      {stories.map((story) => (
        <StoryRow isInSearch={isInSearch} key={story.id} story={story} />
      ))}
    </Box>
  );
};
