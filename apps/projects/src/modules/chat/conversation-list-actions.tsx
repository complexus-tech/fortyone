"use client";

import { NewChatIcon } from "icons";
import { Button, Tooltip } from "ui";
import { ChatNotificationSettings } from "./alerts/settings";
import { ICON_BUTTON_CLASS } from "./chat-styles";

export const ConversationListActions = ({
  onNewMessage,
}: {
  onNewMessage: () => void;
}) => (
  <div className="flex shrink-0 items-center gap-1">
    <ChatNotificationSettings />
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
