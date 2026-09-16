import type { GlassSurfaceProps } from "./glass-surface.types";
import { useState } from "react";
import { StyleSheet, View } from "react-native";
import { Host, Rectangle } from "@expo/ui/swift-ui";
import {
  foregroundStyle,
  frame,
  glassEffect,
} from "@expo/ui/swift-ui/modifiers";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";
import { useLiquidGlass } from "@/hooks/use-liquid-glass";
import { glassFallbackStyle } from "./glass-surface-fallback";

export type { GlassSurfaceProps } from "./glass-surface.types";

export function GlassSurface({
  children,
  cornerRadius = 24,
  selected = false,
  style,
  onLayout,
  ...props
}: GlassSurfaceProps) {
  const { resolvedTheme } = useTheme();
  const [size, setSize] = useState({ width: 0, height: 0 });
  const nativeGlass = useLiquidGlass();
  const theme = themeColors[resolvedTheme];

  return (
    <View
      {...props}
      onLayout={(event) => {
        const { width, height } = event.nativeEvent.layout;
        setSize((current) =>
          current.width === width && current.height === height
            ? current
            : { width, height },
        );
        onLayout?.(event);
      }}
      style={[
        nativeGlass
          ? { borderRadius: cornerRadius }
          : glassFallbackStyle(
              theme,
              resolvedTheme === "dark",
              selected,
              cornerRadius,
            ),
        style,
      ]}
    >
      {nativeGlass && size.width > 0 && size.height > 0 ? (
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
                  glass: {
                    variant: "regular",
                    ...(selected ? { tint: theme.surfaceElevated } : {}),
                  },
                  shape: "roundedRectangle",
                  cornerRadius,
                }),
              ]}
            />
          </Host>
        </View>
      ) : null}
      {nativeGlass && selected ? (
        <View
          pointerEvents="none"
          accessible={false}
          style={[
            StyleSheet.absoluteFill,
            {
              borderRadius: cornerRadius,
              borderWidth: StyleSheet.hairlineWidth,
              borderColor: theme.borderStrong,
            },
          ]}
        />
      ) : null}
      {children}
    </View>
  );
}
