"use client";

import { AiIcon } from "icons";
import { cn } from "lib";
import { Button, Box } from "ui";

type ChatButtonProps = {
  onOpen: () => void;
  isOpen: boolean;
};

export const ChatButton = ({ onOpen, isOpen }: ChatButtonProps) => {
  return (
    <Box
      className={cn(
        "fixed bottom-6 right-4 z-50 hidden transition-all duration-500 ease-in-out md:left-1/2 md:right-1/2 md:-translate-x-1/2",
        {
          "-bottom-16": isOpen,
        },
      )}
    >
      <Button
        className="border-[0.5px] border-gray-200/60 px-3 shadow-xl shadow-gray-200 backdrop-blur dark:border-white/10 dark:shadow-none md:pl-5 md:pr-6"
        color="tertiary"
        leftIcon={<AiIcon className="text-dark-50 dark:text-white" />}
        onClick={onOpen}
        rounded="xl"
        size="lg"
      >
        AI Assistant...
      </Button>
    </Box>
  );
};
