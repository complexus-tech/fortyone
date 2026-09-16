import { useId } from "react";
import { StyleSheet, View } from "react-native";
import Svg, { Defs, RadialGradient, Rect, Stop } from "react-native-svg";
import { colors, themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";

// Mirrors the web Maya page wash. The dark secondary token is brand paper,
// whereas the light secondary token is navy.
const WASHES = [
  { name: "secondary", cx: 18, cy: 0, rx: 54, ry: 32, end: 0.76 },
  { name: "info", cx: 52, cy: -4, rx: 46, ry: 28, end: 0.78 },
  { name: "primary", cx: 88, cy: 0, rx: 52, ry: 30, end: 0.76 },
] as const;
const PAINTED_WASHES = [...WASHES].reverse();

export function MayaBackground() {
  const id = useId();
  const { resolvedTheme } = useTheme();
  const dark = resolvedTheme === "dark";
  const washColors = {
    secondary: dark ? themeColors.dark.foreground : colors.secondary,
    info: colors.info,
    primary: colors.primary,
  };
  const opacity = dark
    ? { secondary: 0.05, info: 0.05, primary: 0.06 }
    : { secondary: 0.07, info: 0.06, primary: 0.07 };
  return (
    <View
      style={StyleSheet.absoluteFill}
      pointerEvents="none"
      accessible={false}
      accessibilityElementsHidden
      importantForAccessibility="no-hide-descendants"
    >
      <Svg
        width="100%"
        height="100%"
        viewBox="0 0 100 100"
        preserveAspectRatio="none"
      >
        <Defs>
          {WASHES.map(({ name, cx, cy, rx, ry, end }) => (
            <RadialGradient
              key={name}
              id={`${id}-${name}`}
              cx={cx}
              cy={cy}
              rx={rx}
              ry={ry}
              gradientUnits="userSpaceOnUse"
            >
              <Stop
                offset={0}
                stopColor={washColors[name]}
                stopOpacity={opacity[name]}
              />
              <Stop offset={end} stopColor={washColors[name]} stopOpacity={0} />
            </RadialGradient>
          ))}
        </Defs>
        {/* CSS lists its topmost background first; SVG paints the last shape on top. */}
        {PAINTED_WASHES.map(({ name }) => (
          <Rect
            key={name}
            width={100}
            height={100}
            fill={`url(#${id}-${name})`}
          />
        ))}
      </Svg>
    </View>
  );
}
