import Svg, { Path } from "react-native-svg";
import { useTheme } from "@/hooks/theme";
import { themeColors } from "@/constants/colors";

/** Geometry shared with packages/icons/src/tags.tsx. */
export function TagsIcon({
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
      strokeLinecap="round"
      strokeLinejoin="round"
      accessible={false}
      accessibilityElementsHidden
      importantForAccessibility="no"
    >
      <Path
        fillRule="evenodd"
        clipRule="evenodd"
        d="M19.4595 1.57208C20.2081 1.74402 20.9664 2.0252 21.4706 2.5294C21.9748 3.0336 22.256 3.79184 22.4279 4.54046C22.6055 5.3138 22.6927 6.19907 22.7288 7.06907C22.8011 8.81159 22.6717 10.5921 22.5881 11.5022C22.5359 12.0702 22.3035 12.5946 21.9407 13.0168C19.2314 16.1691 16.4229 19.0388 13.3564 21.7845C11.9532 23.0408 9.8583 23.0541 8.39717 21.9228C5.98748 20.0571 3.94288 18.0125 2.07714 15.6028C0.945835 14.1417 0.959159 12.0467 2.21547 10.6436C4.96115 7.57712 7.83086 4.76863 10.9832 2.05931C11.4054 1.69644 11.9298 1.46409 12.4977 1.4119C13.4078 1.32827 15.1884 1.19887 16.9309 1.27116C17.8009 1.30726 18.6862 1.39446 19.4595 1.57208ZM19 6.5C19 5.67157 18.3285 5 17.5 5C16.6716 5 16 5.67157 16 6.5C16 7.32843 16.6716 8 17.5 8C18.3285 8 19 7.32843 19 6.5ZM7.70715 13.2929C7.31663 12.9024 6.68346 12.9024 6.29294 13.2929C5.90241 13.6834 5.90241 14.3166 6.29294 14.7071L9.29294 17.7071C9.68346 18.0976 10.3166 18.0976 10.7072 17.7071C11.0977 17.3166 11.0977 16.6834 10.7072 16.2929L7.70715 13.2929Z"
        fill={tint}
      />
    </Svg>
  );
}
