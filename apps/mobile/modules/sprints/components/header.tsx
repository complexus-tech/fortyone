import { Back, ScreenHeader } from "@/components/ui";
import { useLocalSearchParams } from "expo-router";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { useTerminology } from "@/hooks/use-terminology";

export const Header = () => {
  const { teamId } = useLocalSearchParams<{ teamId: string }>();
  const { data: teams = [] } = useTeams();
  const team = teams.find((item) => item.id === teamId);
  const { getTermDisplay } = useTerminology();
  return (
    <ScreenHeader
      title={getTermDisplay("sprintTerm", {
        variant: "plural",
        capitalize: true,
      })}
      subtitle={team?.name}
      leading={<Back />}
      compact
    />
  );
};
