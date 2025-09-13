"use client";
import { BreadCrumbs, Flex, Button } from "ui";
import { GitIcon, SprintsIcon } from "icons";
import { useParams } from "next/navigation";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import { useTeams } from "@/modules/teams/hooks/teams";
import { TeamColor } from "@/components/ui";
import { useTerminology, useUserRole } from "@/hooks";

export const SprintsHeader = () => {
  const { teamId } = useParams<{
    teamId: string;
  }>();
  const { getTermDisplay } = useTerminology();
  const { data: teams = [] } = useTeams();
  const { userRole } = useUserRole();

  const { name, color } = teams.find((team) => team.id === teamId)!;
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
              icon: <SprintsIcon className="h-[1.1rem] w-auto" />,
            },
          ]}
          className="md:hidden"
        />
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
          className="hidden md:flex"
        />
      </Flex>
      <Flex align="center" gap={2}>
        {userRole === "admin" && (
          <Button
            color="tertiary"
            href={`/settings/workspace/teams/${teamId}?tab=automations`}
            leftIcon={<GitIcon />}
            size="sm"
            variant="naked"
          >
            Set up automations
          </Button>
        )}
      </Flex>
    </HeaderContainer>
  );
};
