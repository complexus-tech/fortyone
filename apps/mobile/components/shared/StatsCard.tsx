import React from "react";
import { View, Pressable } from "react-native";
import { Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";

interface StatsCardProps {
  title: string;
  count: number;
  onPress?: () => void;
}

export const StatsCard = ({ title, count, onPress }: StatsCardProps) => {
  return (
    <Pressable
      className="border border-gray-100 rounded-xl px-2.5 py-3 my-1 flex-1"
      onPress={onPress}
    >
      <View className="flex-row justify-between items-center w-full mb-3">
        <Text fontSize="2xl" fontWeight="semibold">
          {count}
        </Text>
        <SymbolView name="plus" size={16} tintColor="#666" />
      </View>
      <Text color="muted">{title}</Text>
    </Pressable>
  );
};
