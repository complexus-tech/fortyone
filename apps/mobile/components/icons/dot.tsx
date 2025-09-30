import React from "react";
import Svg, { Circle } from "react-native-svg";

interface DotProps {
  color: string;
  size?: number;
}

export const Dot = ({ color, size = 12 }: DotProps) => {
  return (
    <Svg width={size} height={size} viewBox="0 0 24 24" fill="none">
      <Circle cx={12} cy={12} r={12} fill={color} />
    </Svg>
  );
};
