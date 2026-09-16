import { Pressable, ScrollView, StyleSheet, View } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { Text } from "@/components/ui";
import { colors, themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks";

export type ApprovalDetail = { label: string; value: string };

export function ApprovalCard({
  title,
  description,
  details,
  destructive = false,
  busy = false,
  onConfirm,
  onCancel,
}: {
  title: string;
  description?: string;
  details: readonly ApprovalDetail[];
  destructive?: boolean;
  busy?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}) {
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];
  return (
    <View
      style={[
        styles.card,
        { backgroundColor: theme.surfaceMuted, borderColor: theme.border },
      ]}
    >
      <View style={styles.heading}>
        <Ionicons
          name={destructive ? "trash-outline" : "checkmark-circle-outline"}
          size={20}
          color={destructive ? colors.danger : theme.foreground}
        />
        <Text fontWeight="bold" style={{ flex: 1 }}>
          {title}
        </Text>
      </View>
      {description ? (
        <Text color="muted" fontSize="sm">
          {description}
        </Text>
      ) : null}
      {details.length > 0 ? (
        <ScrollView
          nestedScrollEnabled
          style={{ maxHeight: 260 }}
          contentContainerStyle={{ gap: 12 }}
        >
          {details.map((detail, index) => (
            <View key={`${index}:${detail.label}`} style={styles.detail}>
              <Text fontSize="sm" color="muted" style={{ flex: 0.4 }}>
                {detail.label}
              </Text>
              <Text fontSize="sm" selectable style={{ flex: 0.6 }}>
                {detail.value}
              </Text>
            </View>
          ))}
        </ScrollView>
      ) : null}
      <View style={styles.actions}>
        <Pressable
          disabled={busy}
          onPress={onCancel}
          accessibilityRole="button"
          accessibilityLabel="Cancel proposed change"
          style={({ pressed }) => [
            styles.button,
            {
              backgroundColor: pressed ? theme.stateActive : theme.accent,
              opacity: busy ? 0.5 : 1,
            },
          ]}
        >
          <Text fontWeight="semibold">Cancel</Text>
        </Pressable>
        <Pressable
          disabled={busy}
          onPress={onConfirm}
          accessibilityRole="button"
          accessibilityLabel={
            destructive ? "Confirm deletion" : "Confirm proposed change"
          }
          accessibilityState={{ disabled: busy, busy }}
          style={({ pressed }) => [
            styles.button,
            {
              backgroundColor: destructive
                ? colors.danger
                : theme.backgroundInverse,
              opacity: pressed || busy ? 0.65 : 1,
            },
          ]}
        >
          <Text
            fontWeight="bold"
            style={{
              color: destructive
                ? colors.dangerForeground
                : theme.foregroundInverse,
            }}
          >
            {busy ? "Please wait…" : destructive ? "Delete" : "Confirm"}
          </Text>
        </Pressable>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    borderRadius: 20,
    borderWidth: StyleSheet.hairlineWidth,
    padding: 16,
    gap: 14,
  },
  heading: { flexDirection: "row", gap: 10, alignItems: "center" },
  detail: { flexDirection: "row", alignItems: "flex-start", gap: 12 },
  actions: { flexDirection: "row", gap: 8, paddingTop: 2 },
  button: {
    flex: 1,
    minHeight: 44,
    borderRadius: 999,
    alignItems: "center",
    justifyContent: "center",
  },
});
