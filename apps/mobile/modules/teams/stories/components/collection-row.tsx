import type { ComponentProps } from "react";
import { memo } from "react";
import { Pressable, View, useColorScheme } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { Text } from "@/components/ui/Text";
import { themeColors } from "@/constants/colors";

type CollectionRowProps = {
  title: string;
  subtitle: string;
  icon: ComponentProps<typeof Ionicons>["name"];
  onPress: () => void;
  accessibilityHint?: string;
};

export const CollectionRow = memo(function CollectionRow({
  title,
  subtitle,
  icon,
  onPress,
  accessibilityHint,
}: CollectionRowProps) {
  const dark = useColorScheme() === "dark";
  const muted = themeColors[dark ? "dark" : "light"].textMuted;
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={[title, subtitle].filter(Boolean).join(", ")}
      accessibilityHint={accessibilityHint}
      onPress={onPress}
      className="active:bg-gray-50 dark:active:bg-dark-200"
      style={{
        flexDirection: "row",
        alignItems: "center",
        gap: 12,
        paddingHorizontal: 20,
        paddingVertical: 14,
        minHeight: 68,
      }}
    >
      <Ionicons name={icon} size={20} color={muted} />
      <View style={{ flex: 1, minWidth: 0, gap: 4 }}>
        <Text numberOfLines={1}>{title}</Text>
        <Text fontSize="xs" color="muted" numberOfLines={1}>
          {subtitle}
        </Text>
      </View>
      <Ionicons name="chevron-forward" size={14} color={muted} />
    </Pressable>
  );
});
