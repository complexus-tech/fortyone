import type { StoriesOptionsSheetProps } from "./stories-options-sheet.shared";
import { useState } from "react";
import { Pressable, ScrollView, StyleSheet, View } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";
import { BottomSheetModal } from "./bottom-sheet-modal";
import { Text } from "./Text";
import {
  getStoriesOptionRows,
  StoriesDisplayOptions,
} from "./stories-options-sheet.shared";

export const StoriesOptionsSheet = (props: StoriesOptionsSheetProps) => {
  const { isOpened, setIsOpened } = props;
  const { resolvedTheme } = useTheme();
  const [activeOptionId, setActiveOptionId] = useState<string | null>(null);
  const rows = getStoriesOptionRows(props);
  const activeOption = rows.find((row) => row.id === activeOptionId);
  const isDark = resolvedTheme === "dark";
  const muted = themeColors[isDark ? "dark" : "light"].textMuted;
  const foreground = themeColors[isDark ? "dark" : "light"].foreground;

  return (
    <>
      <BottomSheetModal
        isOpen={isOpened}
        onClose={() => {
          setActiveOptionId(null);
          setIsOpened(false);
        }}
        spacing={16}
        padding={{ leading: 20, trailing: 20, top: 12, bottom: 24 }}
      >
        <ScrollView
          style={{ flexShrink: 1 }}
          contentContainerStyle={{ gap: 24 }}
          keyboardShouldPersistTaps="handled"
        >
          <View>
            {rows.map((row, index) => (
              <Pressable
                key={row.id}
                accessibilityRole="button"
                accessibilityLabel={`${row.label}, ${row.value}`}
                accessibilityHint="Choose an option"
                onPress={() => setActiveOptionId(row.id)}
                style={({ pressed }) => [
                  styles.row,
                  {
                    opacity: pressed ? 0.6 : 1,
                    borderTopWidth: index > 0 ? StyleSheet.hairlineWidth : 0,
                    borderTopColor: themeColors[resolvedTheme].border,
                  },
                ]}
              >
                <Text style={{ flex: 1 }}>{row.label}</Text>
                <View style={styles.value}>
                  <Text color="muted" style={{ flexShrink: 1 }}>
                    {row.value}
                  </Text>
                  <Ionicons name="chevron-expand" size={16} color={muted} />
                </View>
              </Pressable>
            ))}
          </View>
          <StoriesDisplayOptions {...props} />
        </ScrollView>
      </BottomSheetModal>
      <BottomSheetModal
        isOpen={isOpened && Boolean(activeOption)}
        onClose={() => setActiveOptionId(null)}
        spacing={12}
        padding={{ leading: 20, trailing: 20, top: 12, bottom: 24 }}
      >
        <Text accessibilityRole="header" fontSize="lg" fontWeight="semibold">
          {activeOption?.label}
        </Text>
        <ScrollView style={{ flexShrink: 1 }}>
          {activeOption?.actions.map((action) => (
            <Pressable
              key={action.label}
              accessibilityRole="radio"
              accessibilityState={{ checked: action.selected }}
              onPress={() => {
                setActiveOptionId(null);
                action.onPress();
              }}
              style={({ pressed }) => [
                styles.row,
                { paddingHorizontal: 0, opacity: pressed ? 0.6 : 1 },
              ]}
            >
              <Text style={{ flex: 1 }}>{action.label}</Text>
              {action.selected ? (
                <Ionicons name="checkmark" size={20} color={foreground} />
              ) : null}
            </Pressable>
          ))}
        </ScrollView>
      </BottomSheetModal>
    </>
  );
};

const styles = StyleSheet.create({
  row: {
    minHeight: 52,
    paddingHorizontal: 0,
    paddingVertical: 14,
    flexDirection: "row",
    alignItems: "center",
    gap: 16,
  },
  value: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "flex-end",
    gap: 8,
    maxWidth: "55%",
  },
});
