"use client";
import { BreadCrumbs } from "ui";
import { SprintsIcon } from "icons";
import { useParams } from "next/navigation";
import { HeaderContainer } from "@/components/shared";
import { useTeams } from "@/modules/teams/hooks/teams";
import { NewSprintButton, TeamColor } from "@/components/ui";

export const SprintsHeader = () => {
  const { teamId } = useParams<{
    teamId: string;
  }>();
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
            name: "All sprints",
            icon: <SprintsIcon className="h-[1.1rem] w-auto" />,
          },
        ]}
      />
      <NewSprintButton teamId={teamId} />
    </HeaderContainer>
  );
};
