import type { GlassIconButtonProps } from "./glass-icon-button.types";
import { Pressable, StyleSheet, View } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";

export type { GlassIconButtonProps } from "./glass-icon-button.types";

export function GlassIconButton({
  label,
  onPress,
  disabled = false,
  children,
  icon,
}: GlassIconButtonProps) {
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];

  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={label}
      accessibilityState={{ disabled }}
      disabled={disabled}
      onPress={onPress}
      style={({ pressed }) => [
        styles.control,
        {
          opacity: disabled ? 0.4 : pressed ? 0.65 : 1,
          backgroundColor: theme.surfaceElevated,
          borderColor: theme.border,
        },
      ]}
    >
      <View pointerEvents="none">
        {children ?? (
          <Ionicons
            accessible={false}
            name={icon}
            size={20}
            color={theme.foreground}
          />
        )}
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  control: {
    width: 44,
    height: 44,
    borderRadius: 22,
    borderWidth: StyleSheet.hairlineWidth,
    alignItems: "center",
    justifyContent: "center",
    boxShadow: "0 2px 8px rgba(0, 0, 0, 0.05)",
  },
});
