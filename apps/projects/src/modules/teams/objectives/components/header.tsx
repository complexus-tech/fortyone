"use client";
import { Badge, BreadCrumbs, Button, Flex } from "ui";
import { useState } from "react";
import { PlusIcon, ObjectiveIcon } from "icons";
import { useParams } from "next/navigation";
import { HeaderContainer } from "@/components/shared";
import { NewObjectiveDialog, TeamColor } from "@/components/ui";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";

export const ObjectivesHeader = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: teams = [] } = useTeams();
  const { data: objectives = [] } = useObjectives();
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
              name: "All objectives",
              icon: (
                <ObjectiveIcon className="h-[1.1rem] w-auto" strokeWidth={2} />
              ),
            },
          ]}
        />
        <Badge className="bg-opacity-50" color="tertiary" rounded="full">
          {objectives.length} Objectives
        </Badge>
      </Flex>
      <Flex align="center" gap={2}>
        <Button
          color="tertiary"
          leftIcon={<PlusIcon className="h-[1.1rem] w-auto" />}
          onClick={() => {
            setIsOpen(true);
          }}
          size="sm"
        >
          Create Objective
        </Button>
      </Flex>
      <NewObjectiveDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </HeaderContainer>
  );
};
