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
        "fixed bottom-6 left-1/2 right-1/2 z-50 -translate-x-1/2 transition-all duration-500 ease-in-out",
        {
          "-bottom-16": isOpen,
        },
      )}
    >
      <Button
        className="gap-1.5 border-gray-200/60 px-3 backdrop-blur dark:border-white/10 md:pl-5 md:pr-6"
        color="tertiary"
        leftIcon={
          <AiIcon className="relative h-6 text-dark-50 dark:text-white" />
        }
        onClick={onOpen}
        rounded="xl"
        size="lg"
      >
        <span className="hidden md:inline">AI Assistant</span>
        <span className="inline md:hidden">AI</span>
      </Button>
    </Box>
  );
};
