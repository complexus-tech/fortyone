import type { PropsWithChildren } from "react";
import { useState } from "react";
import { Platform, StyleSheet, View } from "react-native";
import { Host, Rectangle } from "@expo/ui/swift-ui";
import {
  foregroundStyle,
  frame,
  glassEffect,
} from "@expo/ui/swift-ui/modifiers";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";

export function GlassInputSurface({ children }: PropsWithChildren) {
  const { resolvedTheme } = useTheme();
  const [size, setSize] = useState({ width: 0, height: 0 });
  const supportsGlass = Number.parseInt(String(Platform.Version), 10) >= 26;
  const theme = themeColors[resolvedTheme];

  return (
    <View
      onLayout={({ nativeEvent: { layout } }) => {
        setSize((current) =>
          current.width === layout.width && current.height === layout.height
            ? current
            : { width: layout.width, height: layout.height },
        );
      }}
      style={[
        styles.surface,
        !supportsGlass && {
          backgroundColor: theme.surfaceMuted,
          borderColor: theme.borderStrong,
          borderWidth: StyleSheet.hairlineWidth,
        },
      ]}
    >
      {supportsGlass && size.width > 0 && size.height > 0 ? (
        <View
          pointerEvents="none"
          accessibilityElementsHidden
          importantForAccessibility="no-hide-descendants"
          style={StyleSheet.absoluteFill}
        >
          <Host
            pointerEvents="none"
            colorScheme={resolvedTheme}
            ignoreSafeArea="all"
            style={size}
          >
            <Rectangle
              modifiers={[
                foregroundStyle("transparent"),
                frame(size),
                glassEffect({
                  glass: { variant: "regular" },
                  shape: "roundedRectangle",
                  cornerRadius: 24,
                }),
              ]}
            />
          </Host>
        </View>
      ) : null}
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  surface: {
    borderRadius: 24,
  },
});
