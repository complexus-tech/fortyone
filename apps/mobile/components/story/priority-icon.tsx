import React from "react";
import Svg, { Rect, Path } from "react-native-svg";
import { colors } from "@/constants";

type Priority = "Urgent" | "High" | "Medium" | "Low" | "No Priority";

interface PriorityIconProps {
  priority: Priority;
  size?: number;
}

export const PriorityIcon = ({
  priority = "No Priority",
  size = 16,
}: PriorityIconProps) => {
  if (priority === "No Priority") {
    return (
      <Svg
        width={size}
        height={size}
        viewBox="0 0 16 16"
        fill={colors.gray.DEFAULT}
      >
        <Rect
          x="1"
          y="7.25"
          width="3"
          height="1.5"
          rx="0.5"
          opacity="0.9"
          fill="currentColor"
        />
        <Rect
          x="6"
          y="7.25"
          width="3"
          height="1.5"
          rx="0.5"
          opacity="0.9"
          fill="currentColor"
        />
        <Rect
          x="11"
          y="7.25"
          width="3"
          height="1.5"
          rx="0.5"
          opacity="0.9"
          fill="currentColor"
        />
      </Svg>
    );
  }

  if (priority === "Urgent") {
    return (
      <Svg width={20} height={20} viewBox="0 0 24 24" fill="none">
        <Path
          d="M2.5 12C2.5 7.52166 2.5 5.28249 3.89124 3.89124C5.28249 2.5 7.52166 2.5 12 2.5C16.4783 2.5 18.7175 2.5 20.1088 3.89124C21.5 5.28249 21.5 7.52166 21.5 12C21.5 16.4783 21.5 18.7175 20.1088 20.1088C18.7175 21.5 16.4783 21.5 12 21.5C7.52166 21.5 5.28249 21.5 3.89124 20.1088C2.5 18.7175 2.5 16.4783 2.5 12Z"
          stroke={colors.danger}
          strokeWidth="2.5"
          fill={colors.danger}
        />
        <Path
          d="M11.9998 16H12.0088"
          stroke="white"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2.5"
        />
        <Path
          d="M12 13L12 7"
          stroke="white"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2.5"
        />
      </Svg>
    );
  }

  if (priority === "High") {
    return (
      <Svg width={size} height={size} viewBox="0 0 16 16" fill={colors.warning}>
        <Rect x="1" y="8" width="3" height="6" rx="1" fill="currentColor" />
        <Rect x="6" y="5" width="3" height="9" rx="1" fill="currentColor" />
        <Rect x="11" y="2" width="3" height="12" rx="1" fill="currentColor" />
      </Svg>
    );
  }

  if (priority === "Medium") {
    return (
      <Svg width={size} height={size} viewBox="0 0 16 16" fill={colors.success}>
        <Rect x="1" y="8" width="3" height="6" rx="1" fill="currentColor" />
        <Rect x="6" y="5" width="3" height="9" rx="1" fill="currentColor" />
        <Rect
          x="11"
          y="2"
          width="3"
          height="12"
          rx="1"
          fillOpacity="0.4"
          fill="currentColor"
        />
      </Svg>
    );
  }

  if (priority === "Low") {
    return (
      <Svg width={size} height={size} viewBox="0 0 16 16" fill={colors.info}>
        <Rect x="1" y="8" width="3" height="6" rx="1" fill="currentColor" />
        <Rect
          x="6"
          y="5"
          width="3"
          height="9"
          rx="1"
          fillOpacity="0.4"
          fill="currentColor"
        />
        <Rect
          x="11"
          y="2"
          width="3"
          height="12"
          rx="1"
          fillOpacity="0.4"
          fill="currentColor"
        />
      </Svg>
    );
  }

  return null;
};
