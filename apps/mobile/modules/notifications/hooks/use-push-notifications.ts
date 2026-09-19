import { useEffect } from "react";
import * as Notifications from "expo-notifications";
import { useRouter } from "expo-router";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner-native";
import { notificationKeys } from "@/constants/keys";
import { useAuthStore } from "@/store/auth";
import { getObjective } from "@/modules/objectives/queries/get-objective";
import { getSprint } from "@/modules/sprints/queries/get-sprint";
import { readNotification } from "../actions/read";
import { resolveNotificationDestination } from "../utils/destination";
import { registerPushDevice } from "../push/device";

Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldPlaySound: true,
    shouldSetBadge: true,
    shouldShowBanner: true,
    shouldShowList: true,
  }),
});

type PushData = {
  notificationId: string;
  recipientId: string;
  workspaceSlug: string;
  entityType: "story" | "comment" | "objective" | "key_result" | "strategy";
  entityId: string;
};

const readPushData = (value: unknown): PushData | null => {
  if (!value || typeof value !== "object") return null;
  const data = value as Record<string, unknown>;
  const entityTypes = new Set([
    "story",
    "comment",
    "objective",
    "key_result",
    "strategy",
  ]);
  if (
    typeof data.notificationId !== "string" ||
    typeof data.recipientId !== "string" ||
    typeof data.workspaceSlug !== "string" ||
    typeof data.entityId !== "string" ||
    typeof data.entityType !== "string" ||
    !entityTypes.has(data.entityType)
  )
    return null;
  return data as PushData;
};

let lastHandledResponseId: string | null = null;

export const usePushNotifications = () => {
  const router = useRouter();
  const queryClient = useQueryClient();
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const workspace = useAuthStore((state) => state.workspace);
  const userId = useAuthStore((state) => state.userId);

  useEffect(() => {
    if (!isAuthenticated || !workspace || !userId) return;
    void registerPushDevice({ requestPermission: false }).catch(
      () => undefined,
    );

    const openResponse = async (
      response: Notifications.NotificationResponse,
    ) => {
      const responseId = response.notification.request.identifier;
      if (lastHandledResponseId === responseId) return;
      lastHandledResponseId = responseId;
      const data = readPushData(response.notification.request.content.data);
      if (!data) return;
      const openingUser = useAuthStore.getState().userId;
      if (!openingUser || data.recipientId !== openingUser) return;
      if (useAuthStore.getState().workspace !== data.workspaceSlug)
        await useAuthStore.getState().setWorkspace(data.workspaceSlug);
      if (useAuthStore.getState().userId !== openingUser) return;
      const destination = await resolveNotificationDestination(
        {
          id: data.notificationId,
          entityId: data.entityId,
          entityType: data.entityType,
        },
        { loadObjective: getObjective, loadSprint: getSprint },
      );
      if (destination.type === "story") {
        router.push({
          pathname: "/story/[storyId]",
          params: { storyId: destination.storyId },
        });
      } else if (destination.type === "teamStories") {
        router.push(destination.href);
      } else {
        router.push("/inbox");
      }
      void readNotification(data.notificationId).catch(() => undefined);
    };

    const responseSubscription =
      Notifications.addNotificationResponseReceivedListener((response) => {
        void openResponse(response).catch((error: unknown) => {
          toast.error("Could not open notification", {
            description:
              error instanceof Error ? error.message : "Please try again.",
          });
        });
      });
    const receivedSubscription = Notifications.addNotificationReceivedListener(
      () => {
        void queryClient.invalidateQueries({
          queryKey: notificationKeys.all,
        });
      },
    );
    const tokenSubscription = Notifications.addPushTokenListener(() => {
      void registerPushDevice({ requestPermission: false }).catch(
        () => undefined,
      );
    });
    void Notifications.getLastNotificationResponseAsync().then((response) => {
      if (!response) return;
      void openResponse(response)
        .catch(() => undefined)
        .finally(() => Notifications.clearLastNotificationResponseAsync());
    });

    return () => {
      responseSubscription.remove();
      receivedSubscription.remove();
      tokenSubscription.remove();
    };
  }, [isAuthenticated, queryClient, router, userId, workspace]);
};
