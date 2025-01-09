"use client";

import { Box, Container, Flex, Tabs, Text } from "ui";
import { DocsIcon } from "icons";
import { BodyContainer } from "@/components/shared";
import { DocumentRow, Header } from "./components";
import type { DocumentModel } from "./types";

export const DocumentsPage = ({
  documents,
}: {
  documents: DocumentModel[];
}) => {
  const templates = [
    {
      id: 1,
      name: "Blank Document",
    },
    {
      id: 2,
      name: "Requirements Document",
    },
    {
      id: 3,
      name: "Technical Design Document",
    },
    {
      id: 4,
      name: "Objectives and Key Results",
    },
    {
      id: 5,
      name: "Retrospective Document",
    },
    {
      id: 6,
      name: "Retrospective Document",
    },
    {
      id: 7,
      name: "Retrospective Document",
    },
  ];

  return (
    <>
      <Header />
      <BodyContainer className="pt-6">
        <Container className="pb-3">
          <Text as="h3" fontSize="2xl">
            Start from a template
          </Text>
          <Box className="mt-2 overflow-x-auto">
            <Flex className="flex-nowrap p-1" gap={4}>
              {templates.map(({ id, name }) => (
                <Flex
                  align="center"
                  className="w-[200px] shrink-0 cursor-pointer rounded-lg border border-gray-100 bg-gray-50/20 px-3 py-5 shadow-sm transition duration-200 ease-linear dark:border-dark-100 dark:bg-dark-200/40 dark:hover:bg-dark-200/60"
                  gap={3}
                  key={id}
                >
                  <DocsIcon
                    className="h-9 w-auto shrink-0 text-primary"
                    strokeWidth={1.4}
                  />
                  <Text
                    as="h2"
                    className="line-clamp-2"
                    fontSize="lg"
                    fontWeight="medium"
                  >
                    {name}
                  </Text>
                </Flex>
              ))}
            </Flex>
          </Box>
        </Container>

        <Tabs defaultValue="my-docs">
          <Tabs.List className="mb-2.5">
            <Tabs.Tab value="my-docs">My Documents</Tabs.Tab>
            <Tabs.Tab value="favourites">Favourites</Tabs.Tab>
            <Tabs.Tab value="archived">Archive</Tabs.Tab>
          </Tabs.List>
          <Tabs.Panel value="my-docs">
            <Box className="border-t-[0.5px] border-gray-100/60 dark:border-dark-100/80">
              {documents.map(({ id, title }) => (
                <DocumentRow key={id} title={title} />
              ))}
            </Box>
          </Tabs.Panel>
          <Tabs.Panel value="favourites">Favourites</Tabs.Panel>
          <Tabs.Panel value="archived">
            Archive
            <Text>Archive</Text>
          </Tabs.Panel>
        </Tabs>
      </BodyContainer>
    </>
  );
};
