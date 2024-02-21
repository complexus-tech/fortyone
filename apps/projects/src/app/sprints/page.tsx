"use client";

import { Box, Tabs } from "ui";
import { BodyContainer } from "@/components/layout";
import { Header, Project } from "./components";

type Project = {
  id: number;
  code: string;
  lead: string;
  name: string;
  description: string;
  date: string;
};

export default function Page(): JSX.Element {
  const projects: Project[] = [
    {
      id: 1,
      code: "COM-12",
      lead: "John Doe",
      name: "Data migration for Fin connect",
      description: "The quick brown fox jumps over the lazy dog.",
      date: "Sep 27",
    },
    {
      id: 2,
      code: "COM-12",
      lead: "John Doe",
      name: "Complexus data migration",
      description: "Complexus migration to Projects 1.0.0",
      date: "Sep 27",
    },
  ];

  return (
    <>
      <Header />
      <BodyContainer>
        <Tabs defaultValue="all">
          <Box className="sticky top-0 z-10 border-b border-gray-100 bg-white/70 py-3 backdrop-blur dark:border-dark-200 dark:bg-dark-300/60">
            <Tabs.List>
              <Tabs.Tab value="all">All</Tabs.Tab>
              <Tabs.Tab value="active">Active</Tabs.Tab>
              <Tabs.Tab value="backlog">Backlog</Tabs.Tab>
              <Tabs.Tab value="closed">Closed</Tabs.Tab>
            </Tabs.List>
          </Box>
          <Tabs.Panel value="assigned">Tab</Tabs.Panel>
          <Tabs.Panel value="created">Tab</Tabs.Panel>
          <Tabs.Panel value="subscribed">Tab</Tabs.Panel>
        </Tabs>
        {projects.map(({ id, name, description }) => (
          <Project description={description} key={id} name={name} />
        ))}
      </BodyContainer>
    </>
  );
}
