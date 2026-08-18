"use client";
import { BreadCrumbs, Flex } from "ui";
import { useParams } from "next/navigation";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useTerminology } from "@/hooks";

export const SprintsHeader = () => {
  const { teamId } = useParams<{
    teamId: string;
  }>();
  const { getTermDisplay } = useTerminology();
  const { data: teams = [] } = useTeams();

  const selectedTeam = teams.find((team) => team.id === teamId);
  const name = selectedTeam?.name ?? "Team";
  return (
    <HeaderContainer className="justify-between">
      <Flex align="center" gap={2}>
        <MobileMenuButton />
        <BreadCrumbs
          breadCrumbs={[
            {
              name: getTermDisplay("sprintTerm", {
                variant: "plural",
                capitalize: true,
              }),
            },
          ]}
          className="md:hidden"
        />
        <BreadCrumbs
          breadCrumbs={[
            {
              name,
            },
            {
              name: getTermDisplay("sprintTerm", {
                variant: "plural",
                capitalize: true,
              }),
            },
          ]}
          className="hidden md:flex"
        />
      </Flex>
    </HeaderContainer>
  );
};
