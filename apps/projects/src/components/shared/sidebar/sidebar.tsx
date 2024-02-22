"use client";
import { Box, Button, Flex } from "ui";
import { ArrowRightIcon } from "@/components/icons";
import { Header } from "./header";
import { Navigation } from "./navigation";
import { Projects } from "./projects";

export const Sidebar = () => {
  return (
    <Box className="flex h-screen flex-col justify-between px-4 pb-4">
      <Box>
        <Header />
        <Navigation />
        <Projects />
      </Box>
      <Flex justify="between">
        <Button
          className="px-3 text-[0.95rem] font-medium transition duration-200 ease-linear dark:bg-opacity-20 hover:dark:bg-opacity-30"
          rightIcon={<ArrowRightIcon className="h-4 w-auto" />}
          rounded="full"
        >
          Upgrade plan
        </Button>
      </Flex>
    </Box>
  );
};
