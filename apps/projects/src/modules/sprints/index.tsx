"use client";
import { Box, Flex, Text } from "ui";
import { SprintsIcon } from "icons";
import { useParams } from "next/navigation";
import { BodyContainer } from "@/components/shared";
import { NewSprintButton } from "@/components/ui";
import { SprintsHeader } from "./components/header";
import { SprintRow } from "./components/row";
import { SprintsSkeleton } from "./components/sprints-skeleton";
import { useTeamSprints } from "./hooks/team-sprints";

export const SprintsList = () => {
  const { teamId } = useParams<{
    teamId: string;
  }>();

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
                No sprints found
              </Text>
              <Text className="mb-6 max-w-md text-center" color="muted">
                Oops! This team doesn&apos;t have any sprints yet. Create a new
                sprint to get started.
              </Text>
              <Flex gap={2}>
                <NewSprintButton color="tertiary" size="md" teamId={teamId}>
                  Create new sprint
                </NewSprintButton>
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
