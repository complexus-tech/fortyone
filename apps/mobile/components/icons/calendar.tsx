import Svg, { Path } from "react-native-svg";
import { useTheme } from "@/hooks/theme";
import { themeColors } from "@/constants/colors";

/** Same calendar geometry as packages/icons/src/calendar.tsx. */
export function CalendarIcon({
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
      strokeWidth={1.8}
      strokeLinecap="round"
      strokeLinejoin="round"
      accessible={false}
      accessibilityElementsHidden
      importantForAccessibility="no"
    >
      <Path d="M16 2V6M8 2V6" />
      <Path d="M13 4H11C7.229 4 5.343 4 4.172 5.172C3 6.343 3 8.229 3 12V14C3 17.771 3 19.657 4.172 20.828C5.343 22 7.229 22 11 22H13C16.771 22 18.657 22 19.828 20.828C21 19.657 21 17.771 21 14V12C21 8.229 21 6.343 19.828 5.172C18.657 4 16.771 4 13 4Z" />
      <Path d="M3 10H21" />
    </Svg>
  );
}
