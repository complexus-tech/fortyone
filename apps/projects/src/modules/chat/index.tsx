"use client";

import dynamic from "next/dynamic";
import { Button, Skeleton, Text } from "ui";
import { useMayaAssignee } from "@/lib/hooks/members";
import { MobileMenuButton } from "@/components/shared/mobile-menu";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useChatCapabilities } from "@/lib/hooks/workspace/chat-capabilities";

const ChatShell = dynamic(() =>
  import("./chat-shell").then((module) => module.ChatShell),
);

export const ChatPage = () => {
  const chat = useChatCapabilities();
  let availabilityMessage = "Chat is turned off for this workspace.";
  if (chat.isError)
    availabilityMessage = "Chat availability could not be checked.";

  if (!chat.enabled) {
    return (
      <div className="flex h-full flex-col">
        <div className="flex items-center gap-3 px-6 py-4">
          <MobileMenuButton />
          <h1 className="font-semibold">Chat</h1>
        </div>
        {chat.isPending ? (
          <div
            aria-busy="true"
            aria-label="Loading chat"
            className="flex-1 space-y-5 px-6 py-4"
            role="status"
          >
            <Skeleton className="h-10 w-48" />
            <Skeleton className="h-5 w-full max-w-lg" />
            <Skeleton className="h-5 w-2/3 max-w-sm" />
          </div>
        ) : (
          <div className="flex flex-1 flex-col items-center justify-center gap-3 px-6 text-center">
            <Text aria-live="polite" color="muted">
              {availabilityMessage}
            </Text>
            {chat.isError ? (
              <Button
                disabled={chat.isFetching}
                onClick={() => void chat.refetch()}
                variant="outline"
              >
                Try again
              </Button>
            ) : null}
          </div>
        )}
      </div>
    );
  }

  return <EnabledChatPage />;
};

const EnabledChatPage = () => {
  const { data: maya } = useMayaAssignee();
  const { workspaceSlug } = useWorkspacePath();

  return (
    <ChatShell
      key={workspaceSlug}
      mayaAvatarUrl={maya?.avatarUrl}
      navigationControl={<MobileMenuButton />}
    />
  );
};
