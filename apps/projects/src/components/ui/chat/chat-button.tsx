"use client";

import { AiIcon } from "icons";
import { Button, Box } from "ui";

type ChatButtonProps = {
  onOpen: () => void;
};

export const ChatButton = ({ onOpen }: ChatButtonProps) => {
  return (
    <Box className="fixed bottom-8 right-8 z-50">
      <Button
        className="gap-1.5 border-0 bg-gradient-to-r from-primary to-secondary/70 dark:bg-primary dark:text-white md:pl-4 md:pr-5"
        leftIcon={
          <AiIcon className="relative -top-[2px] h-6 text-white dark:text-white" />
        }
        onClick={onOpen}
        rounded="full"
        size="lg"
      >
        Plan with Maya
      </Button>
    </Box>
  );
};
