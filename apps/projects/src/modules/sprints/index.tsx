"use client";
import { Box, Flex, Text, Button } from "ui";
import { SprintsIcon, GitIcon } from "icons";
import { useParams } from "next/navigation";
import { BodyContainer } from "@/components/shared";
import { NewSprintButton } from "@/components/ui";
import { useTerminology, useUserRole } from "@/hooks";
import { SprintsHeader } from "./components/header";
import { SprintRow } from "./components/row";
import { SprintsSkeleton } from "./components/sprints-skeleton";
import { useTeamSprints } from "./hooks/team-sprints";

export const SprintsList = () => {
  const { teamId } = useParams<{
    teamId: string;
  }>();
  const { userRole } = useUserRole();
  const { getTermDisplay } = useTerminology();

  const { data: sprints = [], isPending } = useTeamSprints(teamId);
  if (isPending) {
    return <SprintsSkeleton />;
  }

  return (
    <>
      <SprintsHeader />
      <BodyContainer>
        {sprints.length === 0 && (
          <Box className="flex h-[70dvh] items-center justify-center">
            <Box className="flex flex-col items-center">
              <SprintsIcon className="h-20 w-auto" strokeWidth={1.3} />
              <Text className="mb-6 mt-8" fontSize="3xl">
                No {getTermDisplay("sprintTerm", { variant: "plural" })} found
              </Text>
              <Text className="mb-6 max-w-md text-center" color="muted">
                Oops! This team doesn&apos;t have any{" "}
                {getTermDisplay("sprintTerm")} yet.{" "}
                {userRole !== "admin" &&
                  `Ask an admin to set up ${getTermDisplay("sprintTerm")} automations.`}
              </Text>
              <Flex gap={2}>
                {userRole === "member" && (
                  <NewSprintButton color="tertiary" teamId={teamId}>
                    Create new {getTermDisplay("sprintTerm")}
                  </NewSprintButton>
                )}
                {userRole === "admin" && (
                  <Button
                    color="tertiary"
                    href={`/settings/workspace/teams/${teamId}?tab=automations`}
                    leftIcon={<GitIcon />}
                    size="sm"
                  >
                    Set up automations
                  </Button>
                )}
              </Flex>
            </Box>
          </Box>
        )}
        {sprints.map((sprint) => (
          <SprintRow key={sprint.id} {...sprint} />
        ))}
      </BodyContainer>
    </>
  );
};
