"use client";
import { Box, Tabs } from "ui";
import { BodyContainer } from "@/components/shared";
import { SprintRowsHeader } from "./components/rows-header";
import { SprintRow } from "./components/row";
import { SprintsHeader } from "./components/header";
import { Sprint } from "@/modules/sprints/types";

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
            {sprints.map(({ id, name }) => (
              <SprintRow key={id} title={name} />
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
