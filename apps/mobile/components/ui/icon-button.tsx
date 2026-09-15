import type { ComponentProps, ReactNode } from "react";
import type { PressableProps, StyleProp, ViewStyle } from "react-native";
import { Pressable, StyleSheet } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { colors, themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";

export type IconButtonProps = {
  label: string;
  onPress: NonNullable<PressableProps["onPress"]>;
  disabled?: boolean;
  selected?: boolean;
  style?: StyleProp<ViewStyle>;
} & (
  | { icon: ComponentProps<typeof Ionicons>["name"]; children?: never }
  | { icon?: never; children: ReactNode }
);

export const IconButton = ({
  icon,
  children,
  label,
  onPress,
  disabled = false,
  selected,
  style,
}: IconButtonProps) => {
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={label}
      accessibilityState={{ disabled, selected }}
      disabled={disabled}
      onPress={onPress}
      style={({ pressed }) => [
        styles.button,
        selected && { backgroundColor: `${colors.primary}14` },
        pressed && {
          backgroundColor: theme.stateActive,
        },
        disabled && styles.disabled,
        style,
      ]}
    >
      {children ?? (
        <Ionicons
          accessible={false}
          name={icon}
          size={22}
          color={selected ? colors.primary : theme.foreground}
        />
      )}
    </Pressable>
  );
};

const styles = StyleSheet.create({
  button: {
    width: 44,
    height: 44,
    borderRadius: 12,
    alignItems: "center",
    justifyContent: "center",
  },
  disabled: { opacity: 0.4 },
});
