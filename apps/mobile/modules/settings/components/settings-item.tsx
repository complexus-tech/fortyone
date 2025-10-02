import React from "react";
import { Pressable } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { Row, Text } from "@/components/ui";
import { colors } from "@/constants";

interface SettingsItemProps {
  title: string;
  value?: string;
  onPress?: () => void;
  showChevron?: boolean;
  destructive?: boolean;
}

export const SettingsItem = ({
  title,
  value,
  onPress,
  showChevron = true,
  destructive = false,
}: SettingsItemProps) => {
  return (
    <Pressable
      style={({ pressed }) => [pressed && { backgroundColor: colors.gray[50] }]}
      onPress={onPress}
    >
      <Row align="center" className="px-4 py-3.5">
        <Text color={destructive ? "primary" : "black"} className="flex-1">
          {title}
        </Text>
        <Row align="center">
          {value && (
            <Text color="muted" className="mr-2">
              {value}
            </Text>
          )}
          {showChevron && (
            <Ionicons
              name="chevron-forward"
              size={16}
              color={colors.gray.DEFAULT}
            />
          )}
        </Row>
      </Row>
    </Pressable>
  );
};
