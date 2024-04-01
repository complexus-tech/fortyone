"use client";
import { Box, Button, Flex } from "ui";
import { ArrowRightIcon } from "icons";
import { Header } from "./header";
import { Navigation } from "./navigation";
import { Teams } from "./teams";

export const Sidebar = () => {
  return (
    <Box className="flex h-screen flex-col justify-between bg-gray-100/50 px-4 pb-4 dark:bg-gradient-to-br dark:from-black dark:to-dark">
      <Box>
        <Header />
        <Navigation />
        <Teams />
      </Box>
      <Flex justify="between">
        <Button
          className="px-3 text-[0.95rem] font-medium transition duration-200 ease-linear dark:bg-opacity-20 hover:dark:bg-opacity-30"
          rightIcon={<ArrowRightIcon className="h-3.5 w-auto" />}
          rounded="full"
        >
          Upgrade plan
        </Button>
      </Flex>
    </Box>
  );
};
