import Constants from "expo-constants";
import * as Notifications from "expo-notifications";
import { Platform } from "react-native";
import { post, request } from "@/lib/http";
import { registerPushCleanup } from "@/lib/push-cleanup";

export type PushPermissionState =
  | "granted"
  | "denied"
  | "undetermined"
  | "unsupported";

type RegisteredDevice = {
  id: string;
  platform: "ios" | "android";
  createdAt: string;
  updatedAt: string;
};

const projectId =
  Constants.expoConfig?.extra?.eas?.projectId ?? Constants.easConfig?.projectId;

let registeredToken: string | null = null;

const configureAndroidChannel = () =>
  Platform.OS === "android"
    ? Notifications.setNotificationChannelAsync("default", {
        name: "Notifications",
        importance: Notifications.AndroidImportance.HIGH,
        vibrationPattern: [0, 250, 250, 250],
        lightColor: "#000000",
      })
    : Promise.resolve(null);

export const getPushPermissionState =
  async (): Promise<PushPermissionState> => {
    if (Platform.OS === "web") return "unsupported";
    const permission = await Notifications.getPermissionsAsync();
    if (permission.granted) return "granted";
    return permission.canAskAgain ? "undetermined" : "denied";
  };

export const registerPushDevice = async ({
  requestPermission,
}: {
  requestPermission: boolean;
}): Promise<PushPermissionState> => {
  if (Platform.OS === "web") return "unsupported";
  await configureAndroidChannel();
  let permission = await Notifications.getPermissionsAsync();
  if (!permission.granted && requestPermission && permission.canAskAgain)
    permission = await Notifications.requestPermissionsAsync();
  if (!permission.granted)
    return permission.canAskAgain ? "undetermined" : "denied";
  if (!projectId)
    throw new Error("The Expo project ID is missing from the app build.");
  const token = (await Notifications.getExpoPushTokenAsync({ projectId })).data;
  await post<
    { token: string; platform: "ios" | "android" },
    { data?: RegisteredDevice }
  >(
    "users/notification-devices",
    { token, platform: Platform.OS === "ios" ? "ios" : "android" },
    { useWorkspace: false },
  );
  registeredToken = token;
  return "granted";
};

export const unregisterPushDevice = async () => {
  if (Platform.OS === "web") return;
  let token = registeredToken;
  if (!token && (await getPushPermissionState()) === "granted" && projectId) {
    token = (await Notifications.getExpoPushTokenAsync({ projectId })).data;
  }
  if (!token) return;
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 2_000);
  try {
    await request(
      "delete",
      "users/notification-devices",
      { token, platform: Platform.OS === "ios" ? "ios" : "android" },
      {
        useWorkspace: false,
        handleUnauthorized: false,
        signal: controller.signal,
      },
    );
  } finally {
    clearTimeout(timeout);
  }
  if (registeredToken === token) registeredToken = null;
};

registerPushCleanup(unregisterPushDevice);
