import React from "react";
import { Pressable } from "react-native";
import { useTheme } from "@/hooks";
import { Octicons, Ionicons } from "@expo/vector-icons";
import { Row, Text } from "@/components/ui";
import { colors } from "@/constants";

interface SettingsItemProps {
  title: string;
  value?: string;
  onPress?: () => void;
  showChevron?: boolean;
  destructive?: boolean;
  asLink?: boolean;
  asOptions?: boolean;
}

export const SettingsItem = ({
  title,
  value,
  onPress,
  showChevron = true,
  destructive = false,
  asLink = false,
  asOptions = false,
}: SettingsItemProps) => {
  const { resolvedTheme } = useTheme();
  return (
    <Pressable
      className="active:bg-gray-50 dark:active:bg-dark-200"
      onPress={onPress}
    >
      <Row asContainer align="center" className="py-4">
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
            <>
              {asOptions ? (
                <Ionicons
                  name="chevron-expand"
                  size={16}
                  color={
                    resolvedTheme === "light"
                      ? colors.gray.DEFAULT
                      : colors.gray[300]
                  }
                />
              ) : (
                <Octicons
                  name={asLink ? "arrow-up-right" : "chevron-right"}
                  size={16}
                  color={
                    resolvedTheme === "light"
                      ? colors.gray.DEFAULT
                      : colors.gray[300]
                  }
                />
              )}
            </>
          )}
        </Row>
      </Row>
    </Pressable>
  );
};
