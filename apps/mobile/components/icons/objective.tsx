import Svg, { Path, Circle } from "react-native-svg";
import { useTheme } from "@/hooks/theme";
import { themeColors } from "@/constants/colors";

/** Geometry shared with packages/icons/src/objective.tsx. */
export function ObjectiveIcon({
  size = 20,
  color,
}: {
  size?: number;
  color?: string;
}) {
  const { resolvedTheme } = useTheme();
  const tint = color ?? themeColors[resolvedTheme].icon;
  return (
    <Svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke={tint}
      strokeWidth={2}
      strokeLinecap="round"
      strokeLinejoin="round"
      accessible={false}
      accessibilityElementsHidden
      importantForAccessibility="no"
    >
      <Path
        d="M18 11L20.3458 8.84853C20.7819 8.44853 21 8.24853 21 8M18 5L20.3458 7.15147C20.7819 7.55147 21 7.75147 21 8M21 8C3 8 3 21 3 21"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <Circle cx="5.5" cy="5.5" r="2.5" />
      <Path d="M13 21L18 16M18 21L13 16" strokeLinecap="round" />
    </Svg>
  );
}
