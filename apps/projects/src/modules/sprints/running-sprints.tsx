"use client";
import { Box, Flex, Text } from "ui";
import { BodyContainer } from "@/components/shared";
import { NewSprintButton } from "@/components/ui";
import { SprintsEmptyIllustration } from "@/components/ui/illustrations/empty-state-illustrations";
import { walkthroughTargets } from "@/shared/walkthrough/targets";
import { SprintRow } from "./components/row";
import { RunningSprintsHeader } from "./components/running-sprints-header";
import { SprintsSkeleton } from "./components/sprints-skeleton";
import { useRunningSprints } from "./hooks/running-sprints";

export const RunningSprintsList = () => {
  const { data: sprints = [], isPending } = useRunningSprints();
  if (isPending) {
    return <SprintsSkeleton />;
  }

  return (
    <>
      <RunningSprintsHeader />
      <BodyContainer data-walkthrough-target={walkthroughTargets.sprintsList}>
        {sprints.length === 0 && (
          <Box className="flex h-[70dvh] items-center justify-center">
            <Box className="flex flex-col items-center">
              <SprintsEmptyIllustration />
              <Text className="mt-8 mb-6" fontSize="3xl">
                No sprints found
              </Text>
              <Text className="mb-6 max-w-md text-center" color="muted">
                Oops! This team doesn&apos;t have any sprints yet. Create a new
                sprint to get started.
              </Text>
              <Flex gap={2}>
                <NewSprintButton color="tertiary" size="md">
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
