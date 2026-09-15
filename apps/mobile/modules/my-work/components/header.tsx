import type { StoriesViewOptions } from "@/types/stories-view-options";
import { ScreenHeader } from "@/components/ui/screen-header";
import { StoryOptionsButton } from "@/modules/stories/components";
import { StoryListActions } from "@/modules/stories/components/story-list-actions";

type HeaderProps = {
  viewOptions: StoriesViewOptions;
  setViewOptions: (options: Partial<StoriesViewOptions>) => void;
  resetViewOptions: () => void;
  onFilters: () => void;
  filterCount: number;
};
export const Header = ({
  onFilters,
  filterCount,
  ...displayProps
}: HeaderProps) => (
  <ScreenHeader
    title="My Work"
    trailing={
      <StoryOptionsButton
        {...displayProps}
        renderTrigger={(onDisplay) => (
          <StoryListActions
            onFilters={onFilters}
            onDisplay={onDisplay}
            filterCount={filterCount}
          />
        )}
      />
    }
  />
);
