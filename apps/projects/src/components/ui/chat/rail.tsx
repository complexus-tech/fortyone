"use client";

import { usePathname } from "next/navigation";
import { Box } from "ui";
import { useMediaQuery } from "@/hooks";
import { useChatContext } from "@/context/chat-context";
import { ChatContent } from "./content";

export const ChatRail = () => {
  const pathname = usePathname();
  const isDesktop = useMediaQuery("(min-width: 768px)");
  const { isOpen } = useChatContext();

  if (!isOpen || !isDesktop || pathname.includes("maya")) {
    return null;
  }

  return (
    <Box className="h-full min-h-0">
      <ChatContent />
    </Box>
  );
};
