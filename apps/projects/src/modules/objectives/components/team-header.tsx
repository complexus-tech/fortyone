"use client";
import { BreadCrumbs, Button, Flex } from "ui";
import { useState } from "react";
import { PlusIcon, ObjectiveIcon } from "icons";
import { useParams } from "next/navigation";
import { HeaderContainer } from "@/components/shared";
import { NewObjectiveDialog, TeamColor } from "@/components/ui";
import { useTeams } from "@/modules/teams/hooks/teams";

export const TeamObjectivesHeader = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: teams = [] } = useTeams();
  const { name, color } = teams.find((team) => team.id === teamId)!;
  const [isOpen, setIsOpen] = useState(false);

  return (
    <HeaderContainer className="justify-between">
      <Flex gap={2}>
        <BreadCrumbs
          breadCrumbs={[
            {
              name,
              icon: <TeamColor color={color} />,
            },
            {
              name: "Objectives",
              icon: <ObjectiveIcon className="h-[1.05rem]" strokeWidth={2} />,
            },
          ]}
        />
      </Flex>
      <Flex align="center" gap={2}>
        <Button
          color="tertiary"
          leftIcon={<PlusIcon className="h-[1.1rem]" />}
          onClick={() => {
            setIsOpen(true);
          }}
          size="sm"
        >
          New Objective
        </Button>
      </Flex>
      <NewObjectiveDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </HeaderContainer>
  );
};
