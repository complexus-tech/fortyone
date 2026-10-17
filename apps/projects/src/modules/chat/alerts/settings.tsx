"use client";

import { useEffect, useId, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { SettingsIcon } from "icons";
import { Button, Input, Popover, Switch, Text, Tooltip } from "ui";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { getDesktopBridge } from "@/shell/desktop/bridge";
import { useChatIdentity } from "../chat-context";
import { ICON_BUTTON_CLASS } from "../chat-styles";
import type { ChatAlertPreferences } from "./preferences";
import {
  CHAT_ALERT_PREFERENCES_CHANGED,
  readChatAlertPreferences,
  saveChatAlertPreferences,
} from "./preferences";
import { enableChatSound, playChatSound } from "./sound";

const ChatAlertControls = () => {
  const { userId, conversationId } = useChatIdentity();
  const { workspaceSlug } = useWorkspacePath();
  const queryClient = useQueryClient();
  const id = useId();
  const [preferences, setPreferences] = useState(() =>
    readChatAlertPreferences(userId, workspaceSlug),
  );
  const [desktopAllowed, setDesktopAllowed] = useState(false);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const bridge = getDesktopBridge();

  useEffect(() => {
    let active = true;
    const refresh = () => {
      setPreferences(readChatAlertPreferences(userId, workspaceSlug));
    };
    window.addEventListener("storage", refresh);
    window.addEventListener(CHAT_ALERT_PREFERENCES_CHANGED, refresh);
    if (bridge?.getPreferences) {
      void bridge
        .getPreferences()
        .then((value) => {
          if (active) setDesktopAllowed(value.notifications);
        })
        .catch((cause: unknown) => {
          if (active)
            setError(
              cause instanceof Error
                ? cause.message
                : "Could not load notification permissions.",
            );
        });
    } else {
      setDesktopAllowed(
        typeof Notification !== "undefined" &&
          Notification.permission === "granted",
      );
    }
    return () => {
      active = false;
      window.removeEventListener("storage", refresh);
      window.removeEventListener(CHAT_ALERT_PREFERENCES_CHANGED, refresh);
    };
  }, [bridge, userId, workspaceSlug]);

  const change = async (patch: Partial<ChatAlertPreferences>) => {
    setPending(true);
    setError(null);
    try {
      if (patch.desktop) {
        if (bridge) {
          if (
            !bridge.requestNotificationPermission ||
            !bridge.updatePreferences
          )
            throw new Error("Update FortyOne to enable desktop notifications.");
          if ((await bridge.requestNotificationPermission()) !== "granted")
            throw new Error(
              "Allow FortyOne notifications in your system settings, then try again.",
            );
          const saved = await bridge.updatePreferences({ notifications: true });
          queryClient.setQueryData(["desktop", "preferences"], saved);
        } else {
          if (typeof Notification === "undefined")
            throw new Error(
              "This browser does not support desktop notifications.",
            );
          if ((await Notification.requestPermission()) !== "granted")
            throw new Error(
              "Allow notifications for this site in your browser settings, then try again.",
            );
        }
        setDesktopAllowed(true);
      }
      if (patch.sound && !(await enableChatSound())) {
        throw new Error("Sound is unavailable in this browser.");
      }
      const next = {
        ...readChatAlertPreferences(userId, workspaceSlug),
        ...patch,
      };
      saveChatAlertPreferences(userId, workspaceSlug, next);
      setPreferences(next);
    } catch (cause) {
      setError(
        cause instanceof Error
          ? cause.message
          : "Could not save chat notification preferences.",
      );
    } finally {
      setPending(false);
    }
  };

  const rows = [
    {
      key: "desktop" as const,
      title: "Desktop notifications",
      checked: preferences.desktop && desktopAllowed,
    },
    {
      key: "sound" as const,
      title: "Message sounds",
      checked: preferences.sound,
    },
    {
      key: "quietHours" as const,
      title: "Quiet hours",
      checked: preferences.quietHours,
    },
  ];

  return (
    <div className="space-y-4">
      <div>
        <Text fontWeight="medium">Chat notifications</Text>
        <Text className="mt-1 text-sm" color="muted">
          New messages in this workspace. These settings apply on this device;
          unread counts always stay visible.
        </Text>
      </div>
      {rows.map((row) => (
        <div className="flex items-center justify-between gap-4" key={row.key}>
          <Text className="text-sm" id={`${id}-${row.key}`}>
            {row.title}
          </Text>
          <Switch
            aria-labelledby={`${id}-${row.key}`}
            checked={row.checked}
            disabled={pending}
            onCheckedChange={(checked) => void change({ [row.key]: checked })}
          />
        </div>
      ))}
      {conversationId ? (
        <div className="flex items-center justify-between gap-4">
          <Text className="text-sm" id={`${id}-mute`}>
            Mute this conversation
          </Text>
          <Switch
            aria-labelledby={`${id}-mute`}
            checked={preferences.mutedConversations.includes(conversationId)}
            disabled={pending}
            onCheckedChange={(checked) => {
              const muted = preferences.mutedConversations.filter(
                (value) => value !== conversationId,
              );
              void change({
                mutedConversations: checked
                  ? [...muted, conversationId]
                  : muted,
              });
            }}
          />
        </div>
      ) : null}
      {preferences.quietHours ? (
        <div className="grid grid-cols-2 gap-3">
          <label className="space-y-1 text-sm" htmlFor={`${id}-start`}>
            From
            <Input
              id={`${id}-start`}
              onChange={(event) => {
                if (event.target.value)
                  void change({ quietStart: event.target.value });
              }}
              type="time"
              value={preferences.quietStart}
            />
          </label>
          <label className="space-y-1 text-sm" htmlFor={`${id}-end`}>
            Until
            <Input
              id={`${id}-end`}
              onChange={(event) => {
                if (event.target.value)
                  void change({ quietEnd: event.target.value });
              }}
              type="time"
              value={preferences.quietEnd}
            />
          </label>
          <Text className="col-span-2 text-xs" color="muted">
            Uses this device’s local time.
          </Text>
        </div>
      ) : null}
      <Button
        disabled={!preferences.sound || pending}
        onClick={() => {
          void enableChatSound()
            .then((enabled) => {
              if (!enabled || !playChatSound())
                setError("Sound is unavailable in this browser.");
            })
            .catch(() => {
              setError("Could not play the message sound.");
            });
        }}
        size="sm"
        variant="outline"
      >
        Preview sound
      </Button>
      {bridge && !bridge.chatNotificationLinks ? (
        <Text className="text-xs" color="muted">
          Update the desktop app for direct chat links and silent banners.
        </Text>
      ) : null}
      {pending ? (
        <Text className="text-sm" color="muted" role="status">
          Saving…
        </Text>
      ) : null}
      {error ? (
        <Text className="text-sm" role="alert">
          {error}
        </Text>
      ) : null}
    </div>
  );
};

export const ChatNotificationSettings = () => (
  <Popover>
    <Tooltip title="Chat notification settings">
      <Popover.Trigger asChild>
        <Button
          aria-label="Chat notification settings"
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
    <Popover.Content
      align="end"
      className="max-h-[80dvh] w-80 overflow-y-auto p-4"
    >
      <ChatAlertControls />
    </Popover.Content>
  </Popover>
);
