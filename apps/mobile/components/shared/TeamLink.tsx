import React from "react";
import { View, Pressable } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { Text } from "@/components/ui";

interface TeamLinkProps {
  name: string;
  color: string;
  onPress?: () => void;
}

export const TeamLink = ({ name, color, onPress }: TeamLinkProps) => {
  return (
    <Pressable
      className="bg-white"
      style={({ pressed }) => [pressed && { backgroundColor: "#F2F2F7" }]}
      onPress={onPress}
    >
      <View className="flex-row items-center justify-between py-3.5 min-h-[44px]">
        <View className="flex-row items-center">
          <View
            className="w-3 h-3 rounded-full mr-2"
            style={{ backgroundColor: color }}
          />
          <Text fontSize="md" color="black">
            {name}
          </Text>
        </View>
        <Ionicons name="chevron-forward" size={16} color="#C7C7CC" />
      </View>
    </Pressable>
  );
};
