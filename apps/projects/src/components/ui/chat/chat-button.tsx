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
        className="gap-1.5 md:px-5 md:pl-4"
        color="invert"
        leftIcon={
          <AiIcon className="relative -top-[2px] h-6 text-white dark:text-dark" />
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
