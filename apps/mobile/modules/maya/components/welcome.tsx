import { Pressable, ScrollView, StyleSheet, View } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { Text } from "@/components/ui";
import { themeColors } from "@/constants/colors";
import { useTheme, useTerminology } from "@/hooks";

export function MayaWelcome({
  disabled,
  onPrompt,
}: {
  disabled: boolean;
  onPrompt: (prompt: string) => void;
}) {
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];
  const { getTermDisplay } = useTerminology();
  const suggestions = [
    {
      icon: "sunny-outline" as const,
      title: "Help me focus",
      prompt: "What should I focus on today?",
    },
    {
      icon: "pencil-outline" as const,
      title: `Create a ${getTermDisplay("storyTerm")}`,
      prompt: `Help me create a ${getTermDisplay("storyTerm")}.`,
    },
    {
      icon: "chatbox-ellipses-outline" as const,
      title: "Catch me up",
      prompt:
        "Give me a concise update on my assigned work and anything overdue.",
    },
  ];
  return (
    <ScrollView
      style={styles.scroll}
      contentContainerStyle={styles.welcome}
      keyboardShouldPersistTaps="handled"
      keyboardDismissMode="interactive"
      showsVerticalScrollIndicator={false}
    >
      <View style={styles.spacer} />
      <View style={styles.suggestions}>
        {suggestions.map((suggestion) => (
          <Pressable
            key={suggestion.title}
            accessibilityRole="button"
            accessibilityState={{ disabled }}
            disabled={disabled}
            onPress={() => onPrompt(suggestion.prompt)}
            style={({ pressed }) => [
              styles.suggestion,
              {
                backgroundColor: pressed ? theme.stateActive : "transparent",
                opacity: disabled ? 0.5 : 1,
              },
            ]}
          >
            <Ionicons
              name={suggestion.icon}
              size={22}
              color={theme.textMuted}
            />
            <Text
              color="muted"
              style={styles.suggestionTitle}
              numberOfLines={1}
            >
              {suggestion.title}
            </Text>
          </Pressable>
        ))}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  scroll: { flex: 1, minHeight: 0 },
  spacer: { flexGrow: 1 },
  welcome: {
    flexGrow: 1,
    paddingHorizontal: 16,
    paddingTop: 16,
    paddingBottom: 8,
  },
  suggestions: { width: "100%", gap: 4 },
  suggestion: {
    minHeight: 44,
    borderRadius: 24,
    paddingHorizontal: 8,
    paddingVertical: 10,
    flexDirection: "row",
    gap: 14,
    alignItems: "center",
  },
  suggestionTitle: { flex: 1, fontSize: 17, lineHeight: 24, fontWeight: "400" },
});
