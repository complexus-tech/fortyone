"use client";
import { BreadCrumbs, Button } from "ui";
import { useState } from "react";
import { PlusIcon, ObjectiveIcon } from "icons";
import { HeaderContainer } from "@/components/shared";
import { NewObjectiveDialog, TeamColor } from "@/components/ui";
import { useParams } from "next/navigation";
import { useTeams } from "@/modules/teams/hooks/teams";

export const ObjectivesHeader = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: teams = [] } = useTeams();
  const { name, color } = teams.find((team) => team.id === teamId)!;
  const [isOpen, setIsOpen] = useState(false);
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name,
            icon: <TeamColor color={color} />,
          },
          {
            name: "All objectives",
            icon: <ObjectiveIcon className="h-[1.1rem] w-auto" />,
          },
        ]}
      />
      <Button
        color="tertiary"
        variant="naked"
        leftIcon={<PlusIcon className="h-[1.1rem] w-auto" />}
        onClick={() => {
          setIsOpen(true);
        }}
        size="sm"
      >
        Create Objective
      </Button>
      <NewObjectiveDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </HeaderContainer>
  );
};
