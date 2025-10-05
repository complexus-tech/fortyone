import React from "react";
import { Pressable } from "react-native";
import { useColorScheme } from "nativewind";
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
  const { colorScheme } = useColorScheme();
  return (
    <Pressable
      className="active:bg-gray-50 dark:active:bg-dark-200"
      onPress={onPress}
    >
      <Row asContainer align="center" className="py-3.5">
        <Text color={destructive ? "primary" : undefined} className="flex-1">
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
              color={
                colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
              }
            />
          )}
        </Row>
      </Row>
    </Pressable>
  );
};
