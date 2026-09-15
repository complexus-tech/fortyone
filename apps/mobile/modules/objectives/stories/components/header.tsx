import type { StoriesViewOptions } from "@/types/stories-view-options";
import type { Objective } from "../../types";
import { Back, ScreenHeader } from "@/components/ui";
import { StoryOptionsButton } from "@/modules/stories/components";
import { useTerminology } from "@/hooks/use-terminology";
import { formatContextDates } from "@/modules/teams/stories/context-dates";

type HeaderProps = {
  objective?: Objective;
  viewOptions: StoriesViewOptions;
  setViewOptions: (options: Partial<StoriesViewOptions>) => void;
  resetViewOptions: () => void;
};

export const Header = ({ objective, ...props }: HeaderProps) => {
  const { getTermDisplay } = useTerminology();
  return (
    <ScreenHeader
      title={
        objective?.name || getTermDisplay("objectiveTerm", { capitalize: true })
      }
      subtitle={
        objective
          ? `${getTermDisplay("storyTerm", { variant: "plural", capitalize: true })} · ${formatContextDates(objective.startDate, objective.endDate)}`
          : undefined
      }
      leading={<Back />}
      trailing={<StoryOptionsButton {...props} />}
      compact
    />
  );
};
