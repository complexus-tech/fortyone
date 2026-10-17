"use client";

import Link from "next/link";
import { NewChatIcon, SettingsIcon } from "icons";
import { Button, Popover, Text, Tooltip } from "ui";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { ICON_BUTTON_CLASS } from "./chat-styles";

const ChatSettings = () => {
  const { workspaceSlug } = useWorkspacePath();
  return (
    <Popover>
      <Tooltip title="Chat settings">
        <Popover.Trigger asChild>
          <Button
            aria-label="Chat settings"
            asIcon
            className={ICON_BUTTON_CLASS}
            color="tertiary"
            size="sm"
            type="button"
            variant="naked"
          >
            <SettingsIcon className="size-4" />
          </Button>
        </Popover.Trigger>
      </Tooltip>
      <Popover.Content align="end" className="w-72 p-4">
        <Text fontWeight="medium">Notifications</Text>
        <Text className="mt-2 leading-6" color="muted">
          Manage how FortyOne keeps you informed.
        </Text>
        <Link
          className="text-primary mt-3 inline-block underline underline-offset-2"
          href={`/${workspaceSlug}/settings/account/notifications`}
        >
          Notification settings
        </Link>
      </Popover.Content>
    </Popover>
  );
};

export const ConversationListActions = ({
  onNewMessage,
}: {
  onNewMessage: () => void;
}) => (
  <div className="flex shrink-0 items-center gap-1">
    <ChatSettings />
    <Tooltip title="New message">
      <Button
        aria-label="New message"
        asIcon
        className={ICON_BUTTON_CLASS}
        color="tertiary"
        onClick={onNewMessage}
        size="sm"
        type="button"
        variant="naked"
      >
        <NewChatIcon className="size-4" />
      </Button>
    </Tooltip>
  </div>
);
