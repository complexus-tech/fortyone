"use client";

import { usePathname, useRouter } from "next/navigation";
import { useHotkeys } from "react-hotkeys-hook";
import { useMediaQuery, useUserRole } from "@/hooks";
import { useChatContext } from "@/context/chat-context";
import { useWorkspacePath } from "@/hooks";
import { ChatButton } from "./chat-button";
import { ChatRail } from "./rail";
import { WorkspaceChatLayout } from "./workspace-chat-layout";

export const Chat = () => {
  const pathname = usePathname();
  const router = useRouter();
  const isDesktop = useMediaQuery("(min-width: 768px)");
  const { userRole } = useUserRole();
  const { closeChat, isOpen, openChat } = useChatContext();
  const { withWorkspace } = useWorkspacePath();
  const isOnMayaPage = pathname.includes("maya");

  useHotkeys("shift+m", () => {
    if (userRole === "guest" || isOnMayaPage) {
      return;
    }

    if (isOpen) {
      closeChat();
      return;
    }

    openChat();
  });

  return (
    <>
      {!isOnMayaPage ? (
        <ChatButton
          isOpen={isOpen}
          onOpen={() => {
            if (isDesktop) {
              openChat();
              return;
            }

            router.push(withWorkspace("/maya"));
          }}
        />
      ) : null}
    </>
  );
};

export { ChatRail, WorkspaceChatLayout };
