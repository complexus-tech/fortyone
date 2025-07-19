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
        "fixed bottom-6 right-6 z-50 hidden transition-all duration-500 ease-in-out md:flex",
        {
          "-bottom-16": isOpen,
          "md:hidden": isOnPage,
        },
      )}
    >
      <Button
        className="border-[0.5px] border-gray-200/80 shadow-xl shadow-gray-200 backdrop-blur dark:border-white/10 dark:shadow-none md:h-[3.5rem] md:pl-5 md:pr-6"
        color="tertiary"
        leftIcon={<AiIcon className="h-7 text-dark-50 dark:text-white" />}
        onClick={onOpen}
        rounded="full"
      >
        AI Assistant...
      </Button>
    </Box>
  );
};
