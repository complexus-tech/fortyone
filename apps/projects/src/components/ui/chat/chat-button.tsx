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
        className="bg-white/70 backdrop-blur dark:bg-dark-200/70 md:px-5"
        color="tertiary"
        leftIcon={<AiIcon />}
        onClick={onOpen}
        rounded="full"
        size="lg"
      >
        Ask Maya
      </Button>
    </Box>
  );
};
