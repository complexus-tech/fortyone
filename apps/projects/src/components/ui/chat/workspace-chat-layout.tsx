"use client";

import type { ReactNode } from "react";
import { usePathname } from "next/navigation";
import { ResizablePanel } from "ui";
import { useMediaQuery } from "@/hooks";
import { useChatContext } from "@/context/chat-context";
import { ChatRail } from "./rail";

export const WorkspaceChatLayout = ({
  children,
}: {
  children: ReactNode;
}) => {
  const pathname = usePathname();
  const isDesktop = useMediaQuery("(min-width: 768px)");
  const { isOpen } = useChatContext();
  const isDocked = isOpen && isDesktop && !pathname.includes("maya");

  if (!isDocked) {
    return <>{children}</>;
  }

  return (
    <ResizablePanel autoSaveId="workspace:chat:rail" direction="horizontal">
      <ResizablePanel.Panel defaultSize={72} minSize={45}>
        {children}
      </ResizablePanel.Panel>
      <ResizablePanel.Handle />
      <ResizablePanel.Panel defaultSize={28} maxSize={40} minSize={24}>
        <ChatRail />
      </ResizablePanel.Panel>
    </ResizablePanel>
  );
};
