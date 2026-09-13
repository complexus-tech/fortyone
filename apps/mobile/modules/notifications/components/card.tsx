import React, { useRef, useState } from "react";
import { Pressable } from "react-native";
import { Row, Col, Text, Avatar } from "@/components/ui";
import { colors } from "@/constants";
import type { AppNotification } from "../types";
import { useQueryClient, type QueryClient } from "@tanstack/react-query";
import { openBrowserAsync } from "expo-web-browser";
import { toast } from "sonner-native";
import { objectiveKeys, sprintKeys } from "@/constants/keys";
import { useAuthStore } from "@/store/auth";
import { getApplicationURL } from "@/lib/http/config";
import { getObjective } from "@/modules/objectives/queries/get-objective";
import { getSprint } from "@/modules/sprints/queries/get-sprint";
import { resolveNotificationDestination } from "../utils/destination";
import { renderTemplate, renderTemplateJSX } from "../utils/render-template";
import { useRouter } from "expo-router";
import { Dot } from "@/components/icons";
import { useTerminology } from "@/hooks";
import { useReadNotificationMutation } from "../hooks";

const openNotificationDestination = async (
  notification: Pick<AppNotification, "id" | "entityId" | "entityType">,
  client: QueryClient,
  router: ReturnType<typeof useRouter>,
  workspace: string | null,
  sessionEpoch: number,
) => {
  if (!workspace)
    throw new Error("Choose a workspace to open this notification.");
  const destination = await resolveNotificationDestination(notification, {
    workspace,
    applicationURL: getApplicationURL().toString(),
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
    case "objective":
      router.push({
        pathname: "/team/[teamId]/objectives/[objectiveId]",
        params: {
          teamId: destination.teamId,
          objectiveId: destination.objectiveId,
        },
      });
      break;
    case "sprint":
      router.push({
        pathname: "/team/[teamId]/sprints/[sprintId]",
        params: { teamId: destination.teamId, sprintId: destination.sprintId },
      });
      break;
    case "web":
      await openBrowserAsync(destination.url);
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
  return `${Math.floor(diffInMinutes / 1440)}d`;
};

type NotificationCardProps = AppNotification & {
  index: number;
};

export const NotificationCard = ({
  id,
  title,
  message,
  entityId,
  entityType,
  readAt,
  createdAt,
  actor,
}: NotificationCardProps) => {
  const router = useRouter();
  const { getTermDisplay } = useTerminology();
  const client = useQueryClient();
  const openingRef = useRef(false);
  const [isOpening, setIsOpening] = useState(false);
  // Mark read only after opening a valid destination.
  const { mutate: readNotification } = useReadNotificationMutation();
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
  const nativeDestination = ["story", "objective", "sprint"].includes(
    entityType,
  );

  const handlePress = () => {
    if (openingRef.current) return;
    openingRef.current = true;
    setIsOpening(true);
    const { workspace, sessionEpoch } = useAuthStore.getState();
    void openNotificationDestination(
      { id, entityId, entityType },
      client,
      router,
      workspace,
      sessionEpoch,
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

  return (
    <Pressable
      className="py-3.5 px-4 active:bg-gray-50 dark:active:bg-dark-200"
      onPress={() => {
        void handlePress();
      }}
      disabled={isOpening}
      accessibilityRole="button"
      accessibilityLabel={`${isUnread ? "Unread. " : ""}${title}. ${spokenMessage}`}
      accessibilityHint={
        nativeDestination
          ? "Opens notification details"
          : "Opens the FortyOne website"
      }
      accessibilityState={{ busy: isOpening, disabled: isOpening }}
    >
      <Row align="center" gap={2}>
        <Avatar
          name={actor?.fullName || actor?.username || "Someone"}
          src={actor?.avatarUrl}
          className="shrink-0 relative top-0.5"
          size="lg"
        />
        <Col flex={1} className="gap-1">
          <Row justify="between" align="center" gap={2}>
            <Text
              color={isUnread ? undefined : "muted"}
              className="flex-1 mr-2"
              fontWeight={isUnread ? "medium" : undefined}
              numberOfLines={1}
            >
              {title}
            </Text>
            <Text
              fontSize="sm"
              color={isUnread ? undefined : "muted"}
              fontWeight={isUnread ? "medium" : undefined}
              className="shrink-0"
            >
              {isOpening ? "Opening…" : formatTimeAgo(createdAt)}
            </Text>
          </Row>
          <Row align="center" justify="between" gap={2}>
            <Text numberOfLines={1} align="center" className="flex-1">
              {renderTemplateJSX(messageWithActor, storyTerm)}
            </Text>
            {isUnread && <Dot color={colors.primary} size={10} />}
          </Row>
        </Col>
      </Row>
    </Pressable>
  );
};
