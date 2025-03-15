"use client";
import { BreadCrumbs } from "ui";
import { SprintsIcon } from "icons";
import { useParams } from "next/navigation";
import { HeaderContainer } from "@/components/shared";
import { useTeams } from "@/modules/teams/hooks/teams";
import { NewSprintButton, TeamColor } from "@/components/ui";
import { useTerminologyDisplay } from "@/hooks";

export const SprintsHeader = () => {
  const { teamId } = useParams<{
    teamId: string;
  }>();
  const { getTermDisplay } = useTerminologyDisplay();
  const { data: teams = [] } = useTeams();

  const { name, color } = teams.find((team) => team.id === teamId)!;
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name,
            icon: <TeamColor color={color} />,
          },
          {
            name: getTermDisplay("sprintTerm", {
              variant: "plural",
              capitalize: true,
            }),
            icon: <SprintsIcon className="h-[1.1rem] w-auto" />,
          },
        ]}
      />
      <NewSprintButton teamId={teamId} />
    </HeaderContainer>
  );
};
