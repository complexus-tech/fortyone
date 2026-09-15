import React, { memo, useRef, useState } from "react";
import { Row, Col, Text, Avatar } from "@/components/ui";
import { colors } from "@/constants";
import type { AppNotification } from "../types";
import { useQueryClient, type QueryClient } from "@tanstack/react-query";
import { Alert } from "react-native";
import { toast } from "sonner-native";
import { objectiveKeys, sprintKeys } from "@/constants/keys";
import { useAuthStore } from "@/store/auth";
import { getObjective } from "@/modules/objectives/queries/get-objective";
import { getSprint } from "@/modules/sprints/queries/get-sprint";
import { resolveNotificationDestination } from "../utils/destination";
import { renderTemplate, renderTemplateJSX } from "../utils/render-template";
import { useRouter } from "expo-router";
import { Dot } from "@/components/icons";
import { useTerminology } from "@/hooks";
import { SwipeableRow } from "@/components/ui/swipeable-row";
import {
  useReadNotificationMutation,
  useMarkUnreadMutation,
  useDeleteMutation,
} from "../hooks";

const openNotificationDestination = async (
  notification: Pick<AppNotification, "id" | "entityId" | "entityType">,
  client: QueryClient,
  router: ReturnType<typeof useRouter>,
  workspace: string | null,
  sessionEpoch: number,
  summary: { title: string; description: string },
) => {
  if (!workspace)
    throw new Error("Choose a workspace to open this notification.");
  const destination = await resolveNotificationDestination(notification, {
    loadObjective: (objectiveId) =>
      client.fetchQuery({
        queryKey: objectiveKeys.detail(objectiveId),
        queryFn: ({ signal }) => getObjective(objectiveId, signal),
        networkMode: "always",
        retry: false,
      }),
    loadSprint: (sprintId) =>
      client.fetchQuery({
        queryKey: sprintKeys.detail(sprintId),
        queryFn: ({ signal }) => getSprint(sprintId, signal),
        networkMode: "always",
        retry: false,
      }),
  });
  if (useAuthStore.getState().sessionEpoch !== sessionEpoch) return false;
  switch (destination.type) {
    case "story":
      router.push({
        pathname: "/story/[storyId]",
        params: { storyId: destination.storyId },
      });
      break;
    case "teamStories":
      router.push(destination.href);
      break;
    case "summary":
      Alert.alert(summary.title, summary.description);
      break;
  }
  return true;
};

const formatTimeAgo = (timestamp: string) => {
  const time = new Date(timestamp);
  if (!Number.isFinite(time.getTime())) return "";
  const diffInMinutes = Math.floor((Date.now() - time.getTime()) / 60000);
  if (diffInMinutes < 1) return "now";
  if (diffInMinutes < 60) return `${diffInMinutes}m`;
  if (diffInMinutes < 1440) return `${Math.floor(diffInMinutes / 60)}h`;
  const days = Math.floor(diffInMinutes / 1440);
  if (days < 30) return `${days}d`;
  if (days < 365) return `${Math.floor(days / 30)}mo`;
  return `${Math.floor(days / 365)}y`;
};

export const NotificationCard = memo(function NotificationCard({
  id,
  title,
  message,
  entityId,
  entityType,
  readAt,
  createdAt,
  actor,
}: AppNotification) {
  const router = useRouter();
  const { getTermDisplay } = useTerminology();
  const client = useQueryClient();
  const openingRef = useRef(false);
  const actionInFlight = useRef(false);
  const [isOpening, setIsOpening] = useState(false);
  // Mark read only after opening a valid destination.
  const readMutation = useReadNotificationMutation();
  const { mutate: readNotification } = readMutation;
  const unreadMutation = useMarkUnreadMutation();
  const deleteMutation = useDeleteMutation();
  const sessionEpoch = useAuthStore((state) => state.sessionEpoch);
  const isUpdating =
    readMutation.isPending ||
    unreadMutation.isPending ||
    deleteMutation.isPending;
  const isUnread = !readAt;
  const storyTerm = getTermDisplay("storyTerm");

  const messageWithActor =
    actor && !message?.variables?.actor
      ? {
          ...message,
          template: `{actor} ${message?.template ?? ""}`,
          variables: {
            ...message?.variables,
            actor: { value: actor.fullName || actor.username || "Someone" },
          },
        }
      : message;
  const spokenMessage = renderTemplate(messageWithActor).text;

  const handlePress = () => {
    if (openingRef.current || actionInFlight.current) return;
    openingRef.current = true;
    setIsOpening(true);
    const { workspace, sessionEpoch } = useAuthStore.getState();
    void openNotificationDestination(
      { id, entityId, entityType },
      client,
      router,
      workspace,
      sessionEpoch,
      { title, description: spokenMessage },
    )
      .then((opened) => {
        if (
          opened &&
          isUnread &&
          useAuthStore.getState().sessionEpoch === sessionEpoch
        )
          readNotification(id);
      })
      .catch((error: unknown) => {
        if (useAuthStore.getState().sessionEpoch === sessionEpoch) {
          toast.error("Could not open this notification", {
            description:
              error instanceof Error ? error.message : "Please try again.",
          });
        }
      })
      .finally(() => {
        openingRef.current = false;
        setIsOpening(false);
      });
  };

  const handleAction = (action: "read" | "unread" | "delete") => {
    if (
      openingRef.current ||
      actionInFlight.current ||
      useAuthStore.getState().sessionEpoch !== sessionEpoch
    )
      return;
    actionInFlight.current = true;
    const mutation =
      action === "read"
        ? readMutation
        : action === "unread"
          ? unreadMutation
          : deleteMutation;
    mutation.mutate(id, {
      onSettled: () => {
        actionInFlight.current = false;
      },
    });
  };

  return (
    <SwipeableRow
      className="px-[20px] py-[16px] active:bg-gray-50 dark:active:bg-dark-200"
      onPress={() => {
        void handlePress();
      }}
      disabled={isOpening || isUpdating}
      actionAppearance="flush"
      actionScope={`${sessionEpoch}:${id}`}
      actions={[
        {
          id: isUnread ? "read" : "unread",
          label: isUnread ? "Read" : "Unread",
          accessibilityLabel: isUnread ? "Mark as read" : "Mark as unread",
          icon: isUnread ? "mail-open-outline" : "mail-unread-outline",
          backgroundColor: colors.primary,
          foregroundColor: colors.primaryForeground,
          disabled: isOpening || isUpdating,
          onPress: () => handleAction(isUnread ? "read" : "unread"),
        },
        {
          id: "delete",
          label: "Delete",
          accessibilityLabel: "Delete notification",
          icon: "trash-outline",
          backgroundColor: colors.danger,
          foregroundColor: colors.dangerForeground,
          destructive: true,
          disabled: isOpening || isUpdating,
          onPress: () => handleAction("delete"),
        },
      ]}
      accessibilityRole="button"
      accessibilityLabel={`${isUnread ? "Unread. " : ""}${title}. ${spokenMessage}`}
      accessibilityHint="Opens notification details"
      accessibilityState={{
        busy: isOpening || isUpdating,
        disabled: isOpening || isUpdating,
      }}
    >
      <Row align="start" gap={3}>
        <Avatar
          name={actor?.fullName || actor?.username || "Someone"}
          src={actor?.avatarUrl}
          className="shrink-0 mt-0.5"
          size="md"
          style={{ width: 40, height: 40 }}
        />
        <Col flex={1} className="gap-1">
          <Row justify="between" align="start" gap={2}>
            <Text
              className="flex-1"
              fontWeight="medium"
              color={isUnread ? undefined : "muted"}
              numberOfLines={1}
              ellipsizeMode="tail"
            >
              {title}
            </Text>
            {isUnread ? <Dot color={colors.primary} size={7} /> : null}
          </Row>
          <Text
            fontSize="sm"
            color="muted"
            numberOfLines={1}
            ellipsizeMode="tail"
          >
            {renderTemplateJSX(messageWithActor, storyTerm)}
            {` · ${isOpening ? "Opening…" : formatTimeAgo(createdAt)}`}
          </Text>
        </Col>
      </Row>
    </SwipeableRow>
  );
});
