import React from "react";
import { View } from "react-native";
import { Text } from "@/components/ui";

export const EmptyState = () => {
  return (
    <View className="flex-1 justify-center items-center px-8 py-12">
      <View className="w-16 h-16 rounded-full bg-gray-100 justify-center items-center mb-6">
        <Text fontSize="3xl">🔔</Text>
      </View>

      <Text
        fontSize="xl"
        fontWeight="semibold"
        color="black"
        className="mb-3 text-center"
      >
        No notifications
      </Text>

      <Text fontSize="md" color="muted" className="text-center leading-6">
        You will receive notifications when you are assigned or mentioned in a
        story.
      </Text>
    </View>
  );
};
