import Svg, { Circle, Path } from "react-native-svg";
import { useTheme } from "@/hooks/theme";
import { themeColors } from "@/constants/colors";

// Geometry copied from the corresponding packages/icons/src components.
// The bolder stroke keeps the web family balanced at mobile toolbar sizes.
const GLYPHS = {
  edit: (
    <>
      <Path
        d="M14.0737 3.88545C14.8189 3.07808 15.1915 2.6744 15.5874 2.43893C16.5427 1.87076 17.7191 1.85309 18.6904 2.39232C19.0929 2.6158 19.4769 3.00812 20.245 3.79276C21.0131 4.5774 21.3972 4.96972 21.6159 5.38093C22.1438 6.37312 22.1265 7.57479 21.5703 8.5507C21.3398 8.95516 20.9446 9.33578 20.1543 10.097L10.7506 19.1543C9.25288 20.5969 8.504 21.3182 7.56806 21.6837C6.63212 22.0493 5.6032 22.0224 3.54536 21.9686L3.26538 21.9613C2.63891 21.9449 2.32567 21.9367 2.14359 21.73C1.9615 21.5234 1.98636 21.2043 2.03608 20.5662L2.06308 20.2197C2.20301 18.4235 2.27297 17.5255 2.62371 16.7182C2.97444 15.9109 3.57944 15.2555 4.78943 13.9445L14.0737 3.88545Z"
        stroke="currentColor"
        strokeLinejoin="round"
      />
      <Path d="M13 4L20 11" stroke="currentColor" strokeLinejoin="round" />
      <Path
        d="M14 22H22"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </>
  ),
  user: (
    <>
      <Path d="M20 21C20 18.7328 20 17.5992 19.5929 16.7097C19.1649 15.7746 18.4287 15.0144 17.5071 14.5558C16.6305 14.1196 15.5 14 13.2263 14.0051L10.7728 14C8.49999 14 7.39496 14.1069 6.52924 14.528C5.57859 14.9904 4.82688 15.7651 4.39412 16.7284C4.00001 17.6057 4.00001 18.7371 4 21" />
      <Path d="M16 7C16 9.20914 14.2091 11 12 11C9.79086 11 8 9.20914 8 7C8 4.79086 9.79086 3 12 3C14.2091 3 16 4.79086 16 7Z" />
    </>
  ),
  clock: (
    <>
      <Circle cx="12" cy="12" r="10" />
      <Path d="M12 8V12L14 14" />
    </>
  ),
  sprint: (
    <>
      <Circle cx="12" cy="12" r="10" stroke="currentColor" />
      <Path
        d="M9.5 11.1998V12.8002C9.5 14.3195 9.5 15.0791 9.95576 15.3862C10.4115 15.6932 11.0348 15.3535 12.2815 14.6741L13.7497 13.8738C15.2499 13.0562 16 12.6474 16 12C16 11.3526 15.2499 10.9438 13.7497 10.1262L12.2815 9.32594C11.0348 8.6465 10.4115 8.30678 9.95576 8.61382C9.5 8.92086 9.5 9.6805 9.5 11.1998Z"
        fill="currentColor"
      />
    </>
  ),
  search: (
    <>
      <Path
        d="M17.5 17.5L22 22"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <Path
        d="M20 11C20 6.02944 15.9706 2 11 2C6.02944 2 2 6.02944 2 11C2 15.9706 6.02944 20 11 20C15.9706 20 20 15.9706 20 11Z"
        stroke="currentColor"
        strokeLinejoin="round"
      />
    </>
  ),
  close: (
    <>
      <Path
        d="M19 5L5 19M5 5L19 19"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </>
  ),
  check: (
    <>
      <Path
        d="M5 14L8.5 17.5L19 6.5"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </>
  ),
  chevronDown: (
    <>
      <Path
        d="M5.99977 9.00005L11.9998 15L17.9998 9"
        stroke="currentColor"
        strokeMiterlimit="16"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </>
  ),
  chevronRight: (
    <>
      <Path
        d="M9.00005 6L15 12L9 18"
        stroke="currentColor"
        strokeMiterlimit="16"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </>
  ),
};
export type WebIconName = keyof typeof GLYPHS;

export function WebIcon({
  name,
  size = 20,
  color,
}: {
  name: WebIconName;
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
      color={tint}
      stroke={tint}
      strokeWidth={2.2}
      strokeLinecap="round"
      strokeLinejoin="round"
      accessible={false}
      accessibilityElementsHidden
      importantForAccessibility="no"
    >
      {GLYPHS[name]}
    </Svg>
  );
}
