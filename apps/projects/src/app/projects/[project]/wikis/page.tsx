"use client";

import { Box, Tabs } from "ui";
import { BodyContainer } from "@/components/layout";
import { Header, Wiki } from "./components";

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
          <Box className="sticky top-0 z-10 border-b border-gray-100 bg-white/70 py-3 backdrop-blur dark:border-dark-200 dark:bg-dark-300/80">
            <Tabs.List>
              <Tabs.Tab value="all">Recent</Tabs.Tab>
              <Tabs.Tab value="active">All</Tabs.Tab>
              <Tabs.Tab value="assigned">Favorites</Tabs.Tab>
              <Tabs.Tab value="backlog">Deleted</Tabs.Tab>
              <Tabs.Tab value="closed">Private</Tabs.Tab>
            </Tabs.List>
          </Box>
          <Tabs.Panel value="assigned">Tab</Tabs.Panel>
          <Tabs.Panel value="created">Tab</Tabs.Panel>
          <Tabs.Panel value="subscribed">Tab</Tabs.Panel>
        </Tabs>
        {projects.map(({ id, name }) => (
          <Wiki key={id} name={name} />
        ))}
      </BodyContainer>
    </>
  );
}
