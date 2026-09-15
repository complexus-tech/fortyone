import type { StoriesViewOptions } from "@/types/stories-view-options";
import type { Sprint } from "../../types";
import { Back, ScreenHeader } from "@/components/ui";
import { StoryOptionsButton } from "@/modules/stories/components";
import { useTerminology } from "@/hooks/use-terminology";
import { formatContextDates } from "@/modules/teams/stories/context-dates";

type HeaderProps = {
  sprint?: Sprint;
  viewOptions: StoriesViewOptions;
  setViewOptions: (options: Partial<StoriesViewOptions>) => void;
  resetViewOptions: () => void;
};

export const Header = ({ sprint, ...props }: HeaderProps) => {
  const { getTermDisplay } = useTerminology();
  return (
    <ScreenHeader
      title={sprint?.name || getTermDisplay("sprintTerm", { capitalize: true })}
      subtitle={
        sprint
          ? `${getTermDisplay("storyTerm", { variant: "plural", capitalize: true })} · ${formatContextDates(sprint.startDate, sprint.endDate)}`
          : undefined
      }
      leading={<Back />}
      trailing={<StoryOptionsButton {...props} />}
      compact
    />
  );
};
