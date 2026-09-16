import type { GlassSurfaceProps } from "./glass-surface.types";
import { StyleSheet, View, type ViewStyle } from "react-native";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";

export function GlassSurfaceFallback({
  cornerRadius = 24,
  selected = false,
  style,
  ...props
}: GlassSurfaceProps) {
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];
  const dark = resolvedTheme === "dark";

  return (
    <View
      {...props}
      style={[glassFallbackStyle(theme, dark, selected, cornerRadius), style]}
    />
  );
}

export function glassFallbackStyle(
  theme: {
    surfaceElevated: string;
    surfaceMuted: string;
    foreground: string;
    borderStrong: string;
  },
  dark: boolean,
  selected: boolean,
  cornerRadius: number,
): ViewStyle {
  return {
    borderRadius: cornerRadius,
    backgroundColor: selected ? theme.surfaceElevated : theme.surfaceMuted,
    borderWidth: StyleSheet.hairlineWidth,
    borderColor: selected ? theme.foreground : theme.borderStrong,
    boxShadow: dark
      ? "inset 0 1px 0 rgba(255, 255, 255, 0.10), 0 3px 10px rgba(0, 0, 0, 0.12)"
      : "inset 0 1px 0 rgba(255, 255, 255, 0.85), 0 3px 10px rgba(0, 0, 0, 0.05)",
  };
}
