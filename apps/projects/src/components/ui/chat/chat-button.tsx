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
        className="border-0 bg-gradient-to-r from-primary to-secondary dark:bg-primary dark:text-white md:px-5"
        leftIcon={<AiIcon className="text-white dark:text-white" />}
        onClick={onOpen}
        rounded="full"
        size="lg"
      >
        Ask Maya
      </Button>
    </Box>
  );
};
