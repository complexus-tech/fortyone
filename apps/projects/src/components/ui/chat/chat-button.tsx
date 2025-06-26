"use client";

import { AiIcon } from "icons";
import { Button, Box } from "ui";

type ChatButtonProps = {
  onOpen: () => void;
};

export const ChatButton = ({ onOpen }: ChatButtonProps) => {
  return (
    <Box className="fixed bottom-8 right-8">
      <Button
        className="bg-gray-50/70 backdrop-blur dark:bg-dark-200/70"
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
