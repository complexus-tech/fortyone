import React from "react";
import { View } from "react-native";
import { Text } from "@/components/ui";

interface SectionProps {
  title?: string;
  children: React.ReactNode;
}

export const Section = ({ title, children }: SectionProps) => {
  return (
    <View className="mb-6">
      {title && (
        <Text fontWeight="medium" color="muted" className="mb-1">
          {title}
        </Text>
      )}
      <View className="pl-0.5">{children}</View>
    </View>
  );
};
