import Svg, { Path } from "react-native-svg";
import { useTheme } from "@/hooks/theme";
import { themeColors } from "@/constants/colors";

/** Matches the web FilterIcon geometry with a bolder mobile stroke. */
export function FilterIcon({
  size = 20,
  color,
}: {
  size?: number;
  color?: string;
}) {
  const { resolvedTheme } = useTheme();
  const tint = color ?? themeColors[resolvedTheme].foreground;
  return (
    <Svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      strokeWidth={2.2}
      accessible={false}
      accessibilityElementsHidden
      importantForAccessibility="no"
    >
      <Path
        d="M4 5L20 5"
        stroke={tint}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <Path
        d="M18 12L6 12"
        stroke={tint}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <Path
        d="M8 19L16 19"
        stroke={tint}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </Svg>
  );
}
