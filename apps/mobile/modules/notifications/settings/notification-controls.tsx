import { ActivityIndicator, Pressable, StyleSheet, Switch, View } from "react-native";
import { Text } from "@/components/ui";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks";
import type { NotificationControlsProps } from "./notification-controls.types";

export function NotificationControls({
  enabled,
  enabling,
  testing,
  unavailable,
  onEnabledChange,
  onTest,
}: NotificationControlsProps) {
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];

  return (
    <View style={styles.container}>
      <View
        style={[
          styles.row,
          { backgroundColor: theme.surface, borderColor: theme.border },
        ]}
      >
        <View style={styles.copy}>
          <Text fontWeight="semibold">Allow notifications</Text>
          <Text color="muted" fontSize="sm" numberOfLines={1}>
            Get alerts for important updates
          </Text>
        </View>
        <Switch
          accessibilityLabel="Allow notifications"
          value={enabled}
          disabled={enabling || unavailable}
          onValueChange={onEnabledChange}
        />
      </View>
      {enabled ? (
        <Pressable
          accessibilityRole="button"
          accessibilityLabel="Test notification"
          disabled={testing}
          onPress={onTest}
          style={styles.testLink}
        >
          {testing ? <ActivityIndicator size="small" /> : null}
          <Text color="muted" fontSize="sm">
            {testing ? "Sending test notification…" : "Send test notification"}
          </Text>
        </Pressable>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { gap: 0 },
  row: {
    minHeight: 76,
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderWidth: StyleSheet.hairlineWidth,
    borderRadius: 16,
    flexDirection: "row",
    alignItems: "center",
    gap: 16,
  },
  copy: { flex: 1, minWidth: 0, gap: 3 },
  testLink: {
    minHeight: 32,
    alignSelf: "flex-start",
    marginLeft: 16,
    flexDirection: "row",
    alignItems: "center",
    gap: 8,
  },
});
