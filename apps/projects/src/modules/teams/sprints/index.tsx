"use client";
import { Box, Button, Flex, Tabs, Text } from "ui";
import { BodyContainer } from "@/components/shared";
import { SprintRowsHeader } from "./components/rows-header";
import { SprintRow } from "./components/row";
import { SprintsHeader } from "./components/header";
import { Sprint } from "@/modules/sprints/types";
import { SprintsIcon, StoryMissingIcon } from "icons";

export const SprintsList = ({ sprints }: { sprints: Sprint[] }) => {
  return (
    <>
      <SprintsHeader />
      <Tabs defaultValue="all">
        <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b border-gray-100/60 dark:border-dark-100/40">
          <Tabs.List>
            <Tabs.Tab value="all">All</Tabs.Tab>
            <Tabs.Tab value="active">Active</Tabs.Tab>
            <Tabs.Tab value="upcoming">Upcoming</Tabs.Tab>
            <Tabs.Tab value="completed">Completed</Tabs.Tab>
          </Tabs.List>
        </Box>
        <Tabs.Panel value="all">
          <BodyContainer>
            <SprintRowsHeader />
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
        </Tabs.Panel>
        <Tabs.Panel value="active">Test</Tabs.Panel>
        <Tabs.Panel value="upcoming">Test</Tabs.Panel>
        <Tabs.Panel value="completed">Test</Tabs.Panel>
      </Tabs>
    </>
  );
};
