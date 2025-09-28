import React from "react";
import { View } from "react-native";
import { Text } from "@/components/ui";

interface SectionProps {
  title: string;
  children: React.ReactNode;
}

export const Section = ({ title, children }: SectionProps) => {
  return (
    <View className="mb-6">
      <Text fontWeight="medium" color="muted" className="mb-2">
        {title}
      </Text>
      <View>{children}</View>
    </View>
  );
};
