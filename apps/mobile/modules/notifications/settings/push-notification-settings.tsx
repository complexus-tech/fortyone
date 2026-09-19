import { useCallback, useEffect, useState } from "react";
import { AppState, Linking, ScrollView, StyleSheet, View } from "react-native";
import * as Notifications from "expo-notifications";
import { Ionicons } from "@expo/vector-icons";
import { SafeContainer, Text } from "@/components/ui";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks";
import {
  getPushPermissionState,
  registerPushDevice,
  requestNotificationPermission,
  unregisterPushDevice,
  type PushPermissionState,
} from "../push/device";
import {
  getPushEnabledPreference,
  setPushEnabledPreference,
} from "../push/preference";
import { NotificationControls } from "./notification-controls";
import { PushPreferences } from "./push-preferences";

const errorMessage = (error: unknown) =>
  error instanceof Error ? error.message : "Please try again.";

export function PushNotificationSettings() {
  const [permission, setPermission] =
    useState<PushPermissionState>("undetermined");
  const [enabled, setEnabled] = useState(false);
  const [loading, setLoading] = useState(true);
  const [enabling, setEnabling] = useState(false);
  const [testing, setTesting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];

  const refreshPermission = useCallback(
    () =>
      Promise.all([getPushPermissionState(), getPushEnabledPreference()])
        .then(async ([nextPermission, enabledPreference]) => {
          setError(null);
          setPermission(nextPermission);
          const nextEnabled =
            nextPermission === "granted" && enabledPreference;
          setEnabled(nextEnabled);
          if (nextEnabled) {
            try {
              await registerPushDevice({ requestPermission: false });
            } catch {
              // Local alerts remain available when this development build is
              // signed by an Apple Personal Team without remote push support.
            }
          }
        })
        .catch((cause: unknown) => setError(errorMessage(cause)))
        .finally(() => setLoading(false)),
    [],
  );

  useEffect(() => {
    void Promise.resolve().then(refreshPermission);
    const subscription = AppState.addEventListener("change", (state) => {
      if (state === "active") void refreshPermission();
    });
    return () => subscription.remove();
  }, [refreshPermission]);

  const changeEnabled = (enabled: boolean) => {
    if (enabling || loading || permission === "unsupported") return;
    setError(null);
    if (!enabled) {
      setEnabled(false);
      setEnabling(true);
      void setPushEnabledPreference(false)
        .then(() => unregisterPushDevice())
        .catch(() => undefined)
        .finally(() => setEnabling(false));
      return;
    }
    if (permission === "denied") {
      void setPushEnabledPreference(true);
      void Linking.openSettings().catch((cause: unknown) => {
        setError(errorMessage(cause));
      });
      return;
    }
    setEnabling(true);
    void requestNotificationPermission()
      .then((nextPermission) => {
        setPermission(nextPermission);
        if (nextPermission === "granted") {
          setEnabled(true);
          void setPushEnabledPreference(true)
            .then(() => registerPushDevice({ requestPermission: false }))
            .catch(() => undefined);
        } else {
          setError(
            "Notifications were not allowed. You can enable them in Settings.",
          );
        }
      })
      .catch((cause: unknown) => setError(errorMessage(cause)))
      .finally(() => setEnabling(false));
  };

  const sendTest = () => {
    if (testing || permission !== "granted") return;
    setTesting(true);
    setError(null);
    void Notifications.scheduleNotificationAsync({
      content: {
        title: "FortyOne",
        body: "Notifications are working on this iPhone.",
        sound: true,
      },
      trigger: {
        type: Notifications.SchedulableTriggerInputTypes.TIME_INTERVAL,
        seconds: 1,
      },
    })
      .catch((cause: unknown) => setError(errorMessage(cause)))
      .finally(() => setTesting(false));
  };

  const unavailable = loading || permission === "unsupported";

  return (
    <SafeContainer isFull edges={["bottom"]}>
      <ScrollView
        contentInsetAdjustmentBehavior="automatic"
        showsVerticalScrollIndicator={false}
        contentContainerStyle={styles.content}
      >
        <NotificationControls
          enabled={enabled}
          enabling={enabling}
          testing={testing}
          unavailable={unavailable}
          onEnabledChange={changeEnabled}
          onTest={sendTest}
        />

        {permission === "denied" ? (
          <Text color="muted" fontSize="sm">
            Notifications are off in iPhone Settings. Tap the control above to
            open Settings and allow them.
          </Text>
        ) : null}
        {permission === "unsupported" ? (
          <Text color="muted" fontSize="sm">
            Push notifications are not available on this device.
          </Text>
        ) : null}
        {error ? (
          <View
            accessibilityRole="alert"
            accessibilityLiveRegion="polite"
            style={[
              styles.notice,
              { backgroundColor: theme.surface, borderColor: theme.border },
            ]}
          >
            <Ionicons
              name="alert-circle-outline"
              size={20}
              color={theme.textMuted}
            />
            <Text color="danger" fontSize="sm" style={styles.noticeCopy}>
              {error}
            </Text>
          </View>
        ) : null}
        {enabled ? <PushPreferences /> : null}
      </ScrollView>
    </SafeContainer>
  );
}

const styles = StyleSheet.create({
  content: { paddingHorizontal: 20, paddingTop: 24, paddingBottom: 40, gap: 24 },
  notice: {
    borderWidth: StyleSheet.hairlineWidth,
    borderRadius: 14,
    paddingHorizontal: 14,
    paddingVertical: 12,
    flexDirection: "row",
    alignItems: "flex-start",
    gap: 10,
  },
  noticeCopy: { flex: 1, minWidth: 0 },
});
