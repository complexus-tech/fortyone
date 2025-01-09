"use client";
import { Box, Button, Flex, Tabs, Text } from "ui";
import { SprintsIcon } from "icons";
import { BodyContainer } from "@/components/shared";
import type { Sprint } from "@/modules/sprints/types";
import { SprintRow } from "./components/row";
import { SprintsHeader } from "./components/header";

export const SprintsList = ({ sprints }: { sprints: Sprint[] }) => {
  return (
    <>
      <SprintsHeader />

      <BodyContainer>
        {/* <SprintRowsHeader /> */}
        {sprints.length === 0 && (
          <Box className="flex h-[70vh] items-center justify-center">
            <Box className="flex flex-col items-center">
              <SprintsIcon className="h-20 w-auto" strokeWidth={1.3} />
              <Text className="mb-6 mt-8" fontSize="3xl">
                No sprints found
              </Text>
              <Text className="mb-6 max-w-md text-center" color="muted">
                Oops! This team doesn't have any sprints yet. Create a new
                sprint to get started.
              </Text>
              <Flex gap={2}>
                <Button color="tertiary" size="md">
                  Create new sprint
                </Button>
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
