"use client";
import { BreadCrumbs, Button, Flex } from "ui";
import { useState } from "react";
import { PlusIcon, ObjectiveIcon } from "icons";
import { useParams } from "next/navigation";
import { HeaderContainer } from "@/components/shared";
import { NewObjectiveDialog, TeamColor } from "@/components/ui";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useUserRole, useTerminology } from "@/hooks";

export const TeamObjectivesHeader = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: teams = [] } = useTeams();
  const { name, color } = teams.find((team) => team.id === teamId)!;
  const [isOpen, setIsOpen] = useState(false);
  const { userRole } = useUserRole();
  const { getTermDisplay } = useTerminology();

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
              name: getTermDisplay("objectiveTerm", {
                variant: "plural",
                capitalize: true,
              }),
              icon: <ObjectiveIcon strokeWidth={2} />,
            },
          ]}
        />
      </Flex>
      <Flex align="center" gap={2}>
        <Button
          color="tertiary"
          disabled={userRole === "guest"}
          leftIcon={<PlusIcon className="h-[1.1rem]" />}
          onClick={() => {
            if (userRole !== "guest") {
              setIsOpen(true);
            }
          }}
          size="sm"
        >
          New {getTermDisplay("objectiveTerm", { capitalize: true })}
        </Button>
      </Flex>
      <NewObjectiveDialog isOpen={isOpen} setIsOpen={setIsOpen} />
    </HeaderContainer>
  );
};
