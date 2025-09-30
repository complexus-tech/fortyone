import React from "react";
import Svg, { Circle } from "react-native-svg";

interface StatusDotProps {
  color: string;
  size?: number;
}

export const StatusDot = ({ color, size = 12 }: StatusDotProps) => {
  return (
    <Svg width={size} height={size} viewBox="0 0 24 24" fill="none">
      <Circle cx={12} cy={12} r={12} fill={color} />
    </Svg>
  );
};
