import React from "react";
import { View, Pressable } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { Text } from "@/components/ui";

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
      style={({ pressed }) => [pressed && { backgroundColor: "#F2F2F7" }]}
      onPress={onPress}
    >
      <View className="flex-row items-center px-4 py-3.5 min-h-[44px]">
        <Text color={destructive ? "primary" : "black"} className="flex-1">
          {title}
        </Text>
        <View className="flex-row items-center">
          {value && (
            <Text color="muted" className="mr-2">
              {value}
            </Text>
          )}
          {showChevron && (
            <Ionicons name="chevron-forward" size={16} color="#C7C7CC" />
          )}
        </View>
      </View>
    </Pressable>
  );
};
