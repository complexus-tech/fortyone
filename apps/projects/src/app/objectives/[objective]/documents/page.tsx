"use client";

import { Box, Container, Flex, Tabs, Text } from "ui";
import { ClockIcon, DocsIcon } from "icons";
import { BodyContainer } from "@/components/shared";
import { Header, Doc } from "./components";

type Objective = {
  id: number;
  code: string;
  lead: string;
  name: string;
  description: string;
  date: string;
};

export default function Page(): JSX.Element {
  const objectives: Objective[] = [
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
      description: "Complexus migration to Objectives 1.0.0",
      date: "Sep 27",
    },
  ];

  const categories = [
    {
      id: 1,
      name: "Recent",
    },
    {
      id: 2,
      name: "Favorites",
    },
    {
      id: 3,
      name: "Created by me",
    },
  ];

  return (
    <>
      <Header />
      <BodyContainer>
        <Container className="pb-4 pt-6">
          <Box className="grid grid-cols-3 gap-6">
            {categories.map(({ id, name }) => (
              <Box
                className="rounded-xl border border-gray-100/80 bg-gray-50/20 px-4 py-6 dark:border-dark-100/50 dark:bg-dark-200/50"
                key={id}
              >
                <Flex align="center" className="mb-2" justify="between">
                  <Text
                    as="h2"
                    className="pl-3"
                    fontSize="xl"
                    fontWeight="medium"
                  >
                    {name}
                  </Text>
                  <ClockIcon className="h-6 w-auto" />
                </Flex>
                <Flex
                  align="center"
                  className="rounded-lg px-2 py-2.5 transition duration-200 ease-linear hover:bg-gray-100/50 dark:hover:bg-dark-100"
                  gap={2}
                >
                  <DocsIcon className="h-6 w-auto shrink-0" />
                  <Text color="muted" textOverflow="truncate">
                    Lorem, ipsum dolor sit amet consectetur adipisicing elit.
                    Cumque, quos.
                  </Text>
                </Flex>
                <Flex
                  align="center"
                  className="rounded-lg px-2 py-2.5 transition duration-200 ease-linear dark:hover:bg-dark-100"
                  gap={2}
                >
                  <DocsIcon className="h-6 w-auto shrink-0" />
                  <Text color="muted" textOverflow="truncate">
                    Lorem, ipsum dolor sit amet consectetur adipisicing elit.
                    Cumque, quos.
                  </Text>
                </Flex>
                <Flex
                  align="center"
                  className="rounded-lg px-2 py-2.5 transition duration-200 ease-linear dark:hover:bg-dark-100"
                  gap={2}
                >
                  <DocsIcon className="h-6 w-auto shrink-0" />
                  <Text color="muted" textOverflow="truncate">
                    Lorem, ipsum dolor sit amet consectetur adipisicing elit.
                    Cumque, quos.
                  </Text>
                </Flex>
              </Box>
            ))}
          </Box>
        </Container>

        <Tabs defaultValue="all">
          <Box className="sticky top-0 z-10 border-b border-gray-100 py-3 dark:border-dark-200">
            <Tabs.List>
              <Tabs.Tab value="all">All</Tabs.Tab>
              <Tabs.Tab value="assigned">Private</Tabs.Tab>
              <Tabs.Tab value="backlog">Archived</Tabs.Tab>
              <Tabs.Tab value="closed">Private</Tabs.Tab>
            </Tabs.List>
          </Box>
          <Tabs.Panel value="assigned">Tab</Tabs.Panel>
          <Tabs.Panel value="created">Tab</Tabs.Panel>
          <Tabs.Panel value="subscribed">Tab</Tabs.Panel>
        </Tabs>
        {objectives.map(({ id, name }) => (
          <Doc key={id} name={name} />
        ))}
      </BodyContainer>
    </>
  );
}
