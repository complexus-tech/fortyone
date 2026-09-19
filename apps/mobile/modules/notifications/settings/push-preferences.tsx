import {
  ActivityIndicator,
  Pressable,
  StyleSheet,
  Switch,
  View,
} from "react-native";
import { Text } from "@/components/ui";
import { themeColors } from "@/constants/colors";
import { useFeatures, useTerminology, useTheme } from "@/hooks";
import {
  useNotificationPreferences,
  useUpdatePushPreference,
} from "../hooks/use-notification-preferences";
import type { NotificationType } from "../types";

type PreferenceItem = {
  type: NotificationType;
  title: string;
};

export function PushPreferences() {
  const { resolvedTheme } = useTheme();
  const { getTermDisplay } = useTerminology();
  const features = useFeatures();
  const preferences = useNotificationPreferences();
  const updatePreference = useUpdatePushPreference();
  const theme = themeColors[resolvedTheme];

  const items: PreferenceItem[] = [
    {
      type: "story_update",
      title: `${getTermDisplay("storyTerm", { capitalize: true })} updates`,
    },
    {
      type: "comment_reply",
      title: "Comments",
    },
    {
      type: "mention",
      title: "Mentions",
    },
    {
      type: "story_comment",
      title: `${getTermDisplay("storyTerm", { capitalize: true })} comments`,
    },
    ...(features.objectiveEnabled
      ? [
          {
            type: "objective_update" as const,
            title: `${getTermDisplay("objectiveTerm", { capitalize: true })} updates`,
          },
        ]
      : []),
    ...(features.keyResultEnabled
      ? [
          {
            type: "key_result_update" as const,
            title: `${getTermDisplay("keyResultTerm", { capitalize: true })} updates`,
          },
        ]
      : []),
  ];

  if (preferences.isLoading) {
    return (
      <View style={styles.loading} accessibilityLiveRegion="polite">
        <ActivityIndicator />
        <Text color="muted">Loading notification choices…</Text>
      </View>
    );
  }

  if (preferences.error) {
    return (
      <View style={styles.error}>
        <Text color="danger" accessibilityRole="alert">
          {preferences.error.message}
        </Text>
        <Pressable
          accessibilityRole="button"
          accessibilityLabel="Try loading notification choices again"
          onPress={() => void preferences.refetch()}
        >
          <Text color="muted" fontSize="sm" fontWeight="semibold">
            Try again
          </Text>
        </Pressable>
      </View>
    );
  }

  return (
    <View style={styles.section}>
      <Text color="muted" fontSize="sm" fontWeight="semibold">
        Work updates
      </Text>
      <View
        style={[
          styles.card,
          { backgroundColor: theme.surface, borderColor: theme.border },
        ]}
      >
        {items.map((item, index) => {
          const enabled =
            preferences.data?.preferences[item.type]?.push ?? true;
          const pending =
            updatePreference.isPending &&
            updatePreference.variables?.type === item.type;
          return (
            <View
              key={item.type}
              style={[
                styles.row,
                index > 0 && {
                  borderTopColor: theme.border,
                  borderTopWidth: StyleSheet.hairlineWidth,
                },
              ]}
            >
              <View style={styles.copy}>
                <Text fontWeight="semibold">{item.title}</Text>
              </View>
              <Switch
                accessibilityLabel={`${item.title} push notifications`}
                value={enabled}
                disabled={pending}
                onValueChange={(next) =>
                  updatePreference.mutate({ type: item.type, enabled: next })
                }
              />
            </View>
          );
        })}
      </View>
      {updatePreference.error ? (
        <Text color="danger" fontSize="sm" accessibilityRole="alert">
          {updatePreference.error.message}
        </Text>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  section: { gap: 10 },
  card: { borderWidth: StyleSheet.hairlineWidth, borderRadius: 18 },
  row: {
    minHeight: 56,
    paddingHorizontal: 16,
    paddingVertical: 10,
    flexDirection: "row",
    alignItems: "center",
    gap: 16,
  },
  copy: { flex: 1, minWidth: 0 },
  loading: {
    minHeight: 96,
    alignItems: "center",
    justifyContent: "center",
    gap: 10,
  },
  error: { gap: 8, paddingVertical: 8 },
});
