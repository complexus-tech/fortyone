import type { ReactNode } from "react";
import type { StoriesViewOptions } from "@/types/stories-view-options";
import { useState } from "react";
import { StoriesOptionsSheet } from "@/components/ui";
import { IconButton } from "@/components/ui/icon-button";

type StoryOptionsButtonProps = {
  viewOptions: StoriesViewOptions;
  setViewOptions: (options: Partial<StoriesViewOptions>) => void;
  resetViewOptions: () => void;
  renderTrigger?: (open: () => void) => ReactNode;
};

export const StoryOptionsButton = ({
  viewOptions,
  setViewOptions,
  resetViewOptions,
  renderTrigger,
}: StoryOptionsButtonProps) => {
  const [isOpened, setIsOpened] = useState(false);
  return (
    <>
      {renderTrigger ? (
        renderTrigger(() => setIsOpened(true))
      ) : (
        <IconButton
          icon="options-outline"
          label="View options"
          onPress={() => setIsOpened(true)}
        />
      )}
      <StoriesOptionsSheet
        isOpened={isOpened}
        setIsOpened={setIsOpened}
        viewOptions={viewOptions}
        setViewOptions={setViewOptions}
        resetViewOptions={resetViewOptions}
      />
    </>
  );
};
