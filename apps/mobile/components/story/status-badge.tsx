import React from "react";
import { View, Pressable } from "react-native";
import { Text } from "@/components/ui";

interface StatusBadgeProps {
  status: {
    id: string;
    name: string;
    color: string;
    category?:
      | "unstarted"
      | "started"
      | "completed"
      | "cancelled"
      | "backlog"
      | "paused";
  };
  onPress?: () => void;
  disabled?: boolean;
}

const hexToRgba = (hex: string, alpha: number) => {
  const r = parseInt(hex.slice(1, 3), 16);
  const g = parseInt(hex.slice(3, 5), 16);
  const b = parseInt(hex.slice(5, 7), 16);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
};

export const StatusBadge = ({
  status,
  onPress,
  disabled = true,
}: StatusBadgeProps) => {
  const backgroundColor = hexToRgba(status.color, 0.1);
  const borderColor = hexToRgba(status.color, 0.2);

  const Component = onPress && !disabled ? Pressable : View;

  return (
    <Component
      onPress={onPress}
      style={({ pressed }) => [
        {
          backgroundColor:
            pressed && !disabled
              ? hexToRgba(status.color, 0.15)
              : backgroundColor,
          borderColor,
          borderWidth: 1,
          borderRadius: 6,
          paddingHorizontal: 8,
          paddingVertical: 4,
        },
      ]}
    >
      <Text fontSize="xs" fontWeight="medium" style={{ color: status.color }}>
        {status.name}
      </Text>
    </Component>
  );
};
