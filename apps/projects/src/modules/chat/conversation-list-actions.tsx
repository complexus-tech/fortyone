"use client";

import { useState } from "react";
import Link from "next/link";
import {
  LoadingIcon,
  MenuIcon,
  NewChatIcon,
  NotificationsCheckIcon,
  SettingsIcon,
} from "icons";
import { toast } from "sonner";
import { Button, Menu, Tooltip } from "ui";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useChatAlertPreferences } from "./alerts/use-alert-preferences";
import { useChatIdentity } from "./chat-context";
import { ICON_BUTTON_CLASS } from "./chat-styles";
import { useChatUnread } from "./use-chat-unread";
import { useMarkAllThreadsRead } from "./use-mark-all-threads-read";

export const ConversationListActions = ({
  onNewMessage,
}: {
  onNewMessage: () => void;
}) => {
  const [open, setOpen] = useState(false);
  const { workspaceSlug, withWorkspace } = useWorkspacePath();
  const { userId, conversationId } = useChatIdentity();
  const { preferences, updatePreferences } = useChatAlertPreferences(
    userId,
    workspaceSlug,
  );
  const unread = useChatUnread();
  const markAllRead = useMarkAllThreadsRead();

  const handleMarkAllRead = async () => {
    try {
      await markAllRead.mutateAsync();
      setOpen(false);
      toast.success("All threads marked as read.");
    } catch {
      toast.error("Could not mark threads as read. Please try again.");
    }
  };

  return (
    <div className="flex shrink-0 items-center gap-1">
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
      <Menu onOpenChange={setOpen} open={open}>
        <Tooltip title="Chat menu">
          <Menu.Button>
            <Button
              aria-label="Chat menu"
              asIcon
              className={ICON_BUTTON_CLASS}
              color="tertiary"
              size="sm"
              type="button"
              variant="naked"
            >
              <MenuIcon className="size-4" />
            </Button>
          </Menu.Button>
        </Tooltip>
        <Menu.Items align="end" className="min-w-64">
          <Menu.Group>
            <Menu.Item
              disabled={
                !userId ||
                markAllRead.isPending ||
                unread.data?.unreadThreads === 0
              }
              onSelect={(event) => {
                event.preventDefault();
                void handleMarkAllRead();
              }}
            >
              {markAllRead.isPending ? (
                <LoadingIcon className="size-4 animate-spin" />
              ) : (
                <NotificationsCheckIcon className="size-4" />
              )}
              {markAllRead.isPending
                ? "Marking threads as read…"
                : "Mark all threads as read"}
            </Menu.Item>
            <Menu.Item asChild>
              <Link
                href={withWorkspace("/settings/account/notifications?tab=chat")}
              >
                <SettingsIcon className="size-4" />
                Notification settings
              </Link>
            </Menu.Item>
          </Menu.Group>
          {conversationId ? (
            <>
              <Menu.Separator />
              <Menu.Group>
                <Menu.CheckboxItem
                  checked={preferences.mutedConversations.includes(
                    conversationId,
                  )}
                  onCheckedChange={(checked) => {
                    try {
                      const muted = preferences.mutedConversations.filter(
                        (id) => id !== conversationId,
                      );
                      updatePreferences({
                        mutedConversations: checked
                          ? [...muted, conversationId]
                          : muted,
                      });
                    } catch {
                      toast.error(
                        "Could not update conversation notifications.",
                      );
                    }
                  }}
                >
                  <span className="pl-6">Mute this conversation</span>
                </Menu.CheckboxItem>
              </Menu.Group>
            </>
          ) : null}
        </Menu.Items>
      </Menu>
    </div>
  );
};
