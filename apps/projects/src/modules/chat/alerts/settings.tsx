"use client";

import { useEffect, useId, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Box, Button, Flex, Input, Switch, Text } from "ui";
import { SectionHeader } from "@/components/ui/section-header";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import { getDesktopBridge } from "@/shell/desktop/bridge";
import type { ChatAlertPreferences } from "./preferences";
import { enableChatSound, playChatSound } from "./sound";
import { useChatAlertPreferences } from "./use-alert-preferences";

const ChatAlertControls = ({
  userId,
  workspaceSlug,
}: {
  userId: string;
  workspaceSlug: string;
}) => {
  const queryClient = useQueryClient();
  const id = useId();
  const { preferences, updatePreferences } = useChatAlertPreferences(
    userId,
    workspaceSlug,
  );
  const [desktopAllowed, setDesktopAllowed] = useState(false);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const bridge = getDesktopBridge();

  useEffect(() => {
    let active = true;
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
    };
  }, [bridge]);

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
      updatePreferences(patch);
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
      description: "Show desktop banners for new messages in this workspace.",
      checked: preferences.desktop && desktopAllowed,
    },
    {
      key: "sound" as const,
      title: "Message sounds",
      description: "Play a sound when a new chat message arrives.",
      checked: preferences.sound,
    },
    {
      key: "quietHours" as const,
      title: "Quiet hours",
      description: "Pause chat banners and sounds during a set time.",
      checked: preferences.quietHours,
    },
  ];

  return (
    <Flex direction="column" gap={6}>
      {rows.map((row) => (
        <Flex align="center" gap={4} justify="between" key={row.key}>
          <Box>
            <Text className="font-medium" id={`${id}-${row.key}`}>
              {row.title}
            </Text>
            <Text color="muted" id={`${id}-${row.key}-description`}>
              {row.description}
            </Text>
          </Box>
          <Switch
            aria-describedby={`${id}-${row.key}-description`}
            aria-labelledby={`${id}-${row.key}`}
            checked={row.checked}
            className="shrink-0"
            disabled={pending}
            onCheckedChange={(checked) => void change({ [row.key]: checked })}
          />
        </Flex>
      ))}
      {preferences.quietHours ? (
        <Box className="grid max-w-md grid-cols-2 gap-3">
          <label className="space-y-1" htmlFor={`${id}-start`}>
            From
            <Input
              disabled={pending}
              id={`${id}-start`}
              onChange={(event) => {
                if (event.target.value)
                  void change({ quietStart: event.target.value });
              }}
              type="time"
              value={preferences.quietStart}
            />
          </label>
          <label className="space-y-1" htmlFor={`${id}-end`}>
            Until
            <Input
              disabled={pending}
              id={`${id}-end`}
              onChange={(event) => {
                if (event.target.value)
                  void change({ quietEnd: event.target.value });
              }}
              type="time"
              value={preferences.quietEnd}
            />
          </label>
          <Text className="col-span-2" color="muted">
            Uses this device’s local time.
          </Text>
        </Box>
      ) : null}
      <Button
        className="self-start"
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
      <Text color="muted">
        These settings apply on this device. Unread counts stay visible.
      </Text>
      {bridge && !bridge.chatNotificationLinks ? (
        <Text color="muted">
          Update the desktop app for direct chat links and silent banners.
        </Text>
      ) : null}
      {pending ? (
        <Text color="muted" role="status">
          Saving…
        </Text>
      ) : null}
      {error ? <Text role="alert">{error}</Text> : null}
    </Flex>
  );
};

export const ChatNotificationSettings = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const userId = session?.user.id;

  return (
    <Box className="border-border bg-surface rounded-2xl border">
      <SectionHeader
        description="Choose how new chat messages notify you in this workspace."
        title="Chat Notifications"
      />
      <Box className="p-6">
        {userId && workspaceSlug ? (
          <ChatAlertControls
            key={`${userId}:${workspaceSlug.toLowerCase()}`}
            userId={userId}
            workspaceSlug={workspaceSlug}
          />
        ) : (
          <Text color="muted">Sign in to manage chat notifications.</Text>
        )}
      </Box>
    </Box>
  );
};
