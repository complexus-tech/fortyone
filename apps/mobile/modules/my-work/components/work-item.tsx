import React from "react";
import { View, Pressable } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { Text } from "@/components/ui";

interface WorkItemProps {
  title: string;
  status: "todo" | "in-progress" | "in-review" | "done";
  assignee?: { name: string };
  onPress?: () => void;
}

export const WorkItem = ({
  title,
  status,
  assignee,
  onPress,
}: WorkItemProps) => {
  const getStatusColor = () => {
    switch (status) {
      case "todo":
        return "#6b7280"; // gray
      case "in-progress":
        return "#f59e0b"; // warning
      case "in-review":
        return "#3b82f6"; // blue
      case "done":
        return "#22c55e"; // success
      default:
        return "#6b7280";
    }
  };

  return (
    <Pressable
      style={({ pressed }) => [pressed && { backgroundColor: "#F2F2F7" }]}
      onPress={onPress}
    >
      <View className="flex-row items-center justify-between py-3.5 min-h-[44px]">
        <View className="flex-row items-center flex-1">
          <View
            className="w-3 h-3 rounded-full mr-3"
            style={{ backgroundColor: getStatusColor() }}
          />
          <View className="flex-1">
            <Text fontSize="md" color="black">
              {title}
            </Text>
            {assignee && (
              <Text fontSize="sm" color="muted" className="mt-1">
                Assigned to {assignee.name}
              </Text>
            )}
          </View>
        </View>
        <Ionicons name="chevron-forward" size={16} color="#C7C7CC" />
      </View>
    </Pressable>
  );
};
