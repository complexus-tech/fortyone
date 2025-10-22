import React from "react";
import { View } from "react-native";
import { Text } from "@/components/ui";

interface SettingsSectionProps {
  title?: string;
  children: React.ReactNode;
}

export const SettingsSection = ({ title, children }: SettingsSectionProps) => {
  return (
    <View className="mb-6">
      {title && (
        <Text color="muted" className="mb-2 mx-4.5">
          {title}
        </Text>
      )}
      <View className="flex-1 pl-0.5">{children}</View>
    </View>
  );
};
