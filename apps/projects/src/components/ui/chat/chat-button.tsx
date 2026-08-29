"use client";

import { AiIcon } from "icons";
import { cn } from "lib";
import { usePathname } from "next/navigation";
import { Button, Box } from "ui";

type ChatButtonProps = {
  onOpen: () => void;
  isOpen: boolean;
};

export const ChatButton = ({ onOpen, isOpen }: ChatButtonProps) => {
  const pathname = usePathname();
  const isOnPage = pathname.includes("maya");

  return (
    <Box
      className={cn(
        "fixed right-6 bottom-6 z-50 hidden transition-all duration-500 ease-in-out md:flex",
        {
          "-bottom-16": isOpen,
          "md:hidden": isOnPage,
        },
      )}
    >
      <Button
        className="border-0 px-4 shadow-xl shadow-gray-200 md:h-12 dark:shadow-none"
        color="invert"
        data-chat-button
        leftIcon={<AiIcon className="h-5 text-current dark:text-current" />}
        onClick={onOpen}
        rounded="md"
      >
        Ask Maya AI
      </Button>
    </Box>
  );
};
