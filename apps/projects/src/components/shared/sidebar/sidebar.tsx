"use client";
import { Box, Button, Flex } from "ui";
import { ArrowRightIcon } from "@/components/icons";
import { Header } from "./header";
import { Navigation } from "./navigation";
import { Projects } from "./projects";

export const Sidebar = () => {
  return (
    <Box className="flex h-screen flex-col justify-between border-r border-gray-50 px-4 pb-4 dark:border-dark-100/50">
      <Box>
        <Header />
        <Navigation />
        <Projects />
      </Box>
      <Flex justify="between">
        <Button
          className="px-3 text-[0.95rem] font-medium dark:border-dark-100"
          color="tertiary"
          rightIcon={<ArrowRightIcon className="h-4 w-auto" />}
          rounded="full"
          variant="outline"
        >
          Upgrade plan
        </Button>
      </Flex>
    </Box>
  );
};
