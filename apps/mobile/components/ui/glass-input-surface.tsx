import type { PropsWithChildren } from "react";
import { StyleSheet, View } from "react-native";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";

export function GlassInputSurface({ children }: PropsWithChildren) {
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];
  return (
    <View
      style={[
        styles.surface,
        {
          backgroundColor: theme.surfaceMuted,
          borderColor: theme.borderStrong,
        },
      ]}
    >
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  surface: {
    borderRadius: 24,
    borderWidth: StyleSheet.hairlineWidth,
  },
});
