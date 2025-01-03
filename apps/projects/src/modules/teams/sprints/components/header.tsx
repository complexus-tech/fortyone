"use client";
import { Avatar, BreadCrumbs, Button, Flex } from "ui";
import {
  PlusIcon,
  SearchIcon,
  PreferencesIcon,
  ArrowDownIcon,
  SprintsIcon,
} from "icons";
import { HeaderContainer } from "@/components/shared";
import { useParams } from "next/navigation";
import { useTeams } from "@/modules/teams/hooks/teams";
import { TeamColor } from "@/components/ui";

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
      <Button
        leftIcon={<PlusIcon className="text-white dark:text-gray-200" />}
        size="sm"
      >
        New sprint
      </Button>
    </HeaderContainer>
  );
};
