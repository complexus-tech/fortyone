"use client";

import { useState } from "react";
import Image from "next/image";
import { Box, Button, Flex, Text } from "ui";
import { ImportWizard } from "./components/import-wizard";

const IMPORT_SOURCES = [
  {
    alt: "Jira",
    className: "h-8 w-8",
    height: 32,
    src: "/integrations/jira.svg",
    width: 32,
  },
  {
    alt: "ClickUp",
    className: "h-8 w-8",
    height: 32,
    src: "/integrations/clickup.svg",
    width: 32,
  },
  {
    alt: "monday.com",
    className: "h-7 w-10 object-contain",
    height: 24,
    src: "/integrations/monday.svg",
    width: 40,
  },
  {
    alt: "Asana",
    className: "h-8 w-8",
    height: 32,
    src: "/integrations/asana.svg",
    width: 32,
  },
] as const;

export const WorkspaceImportSettings = ({
  openFromOnboarding = false,
}: {
  openFromOnboarding?: boolean;
}) => {
  const [open, setOpen] = useState(openFromOnboarding);

  return (
    <Box>
      <Text as="h1" className="mb-2 text-2xl font-medium">
        Import work
      </Text>
      <Text className="max-w-3xl leading-6" color="muted">
        Bring an existing backlog into FortyOne without rebuilding it by hand.
        Every import is mapped automatically and shown as a preview before
        anything is created.
      </Text>

      <Box
        aria-labelledby="import-export-heading"
        as="section"
        className="border-border bg-surface mt-6 rounded-2xl border-[0.5px] p-6"
      >
        <Flex gap={3} wrap>
          {IMPORT_SOURCES.map((source) => (
            <Box
              className="bg-surface-muted flex h-12 w-12 items-center justify-center rounded-xl"
              key={source.alt}
              title={source.alt}
            >
              <Image
                alt={source.alt}
                className={source.className}
                height={source.height}
                src={source.src}
                width={source.width}
              />
            </Box>
          ))}
        </Flex>

        <Text
          as="h2"
          className="mt-6 text-xl font-medium"
          id="import-export-heading"
        >
          Import issues from Jira, ClickUp, or anywhere
        </Text>
        <Text className="mt-2 max-w-2xl leading-6" color="muted">
          Upload an export from Jira, ClickUp, monday.com, Asana, or any tool
          that gives you a CSV. Excel files, PDFs, and images work too.
        </Text>

        <Button
          className="mt-5"
          color="invert"
          onClick={() => {
            setOpen(true);
          }}
          size="lg"
        >
          Import
        </Button>
      </Box>

      <ImportWizard onOpenChange={setOpen} open={open} />
    </Box>
  );
};
