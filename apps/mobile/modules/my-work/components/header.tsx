import React from "react";
import { View } from "react-native";
import { Text } from "@/components/ui";

export const Header = () => {
  return (
    <View className="px-4 pb-4">
      <Text fontSize="2xl" fontWeight="semibold">
        My Work
      </Text>
    </View>
  );
};
