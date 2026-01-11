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
        asIcon
        className="border-0 shadow-xl shadow-gray-200 md:h-[3.4rem] dark:shadow-none"
        color="invert"
        data-chat-button
        leftIcon={<AiIcon className="h-7 text-current dark:text-current" />}
        onClick={onOpen}
        rounded="full"
      >
        <span className="sr-only">AI Assistant</span>
      </Button>
    </Box>
  );
};
