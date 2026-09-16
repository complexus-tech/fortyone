import type { StoriesViewOptions } from "@/types/stories-view-options";
import { Back, ScreenHeader } from "@/components/ui";
import { useLocalSearchParams } from "expo-router";
import { StoryOptionsButton } from "@/modules/stories/components";
import { StoryListActions } from "@/modules/stories/components/story-list-actions";
import { useTeams } from "@/modules/teams/hooks/use-teams";

type HeaderProps = {
  viewOptions: StoriesViewOptions;
  setViewOptions: (options: Partial<StoriesViewOptions>) => void;
  resetViewOptions: () => void;
  onFilters: () => void;
  filterCount: number;
  showStoryActions?: boolean;
};

export const Header = ({
  onFilters,
  filterCount,
  showStoryActions = true,
  ...displayProps
}: HeaderProps) => {
  const { teamId } = useLocalSearchParams<{ teamId: string }>();
  const { data: teams = [] } = useTeams();
  const team = teams.find((item) => item.id === teamId);
  return (
    <ScreenHeader
      title={team?.name ?? "Team"}
      leading={<Back />}
      compact
      trailing={
        showStoryActions ? (
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
        ) : undefined
      }
    />
  );
};
