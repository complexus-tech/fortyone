import React from "react";
import { Badge, Text } from "@/components/ui";
import { PriorityIcon } from "@/components/icons";
import { StoryPriority } from "@/modules/stories/types";
import { Pressable } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useColorScheme } from "nativewind";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { PropertyBottomSheet } from "./property-bottom-sheet";

const Item = ({
  priority,
  onPress,
  isSelected,
}: {
  priority: StoryPriority;
  onPress: () => void;
  isSelected: boolean;
}) => {
  const { colorScheme } = useColorScheme();
  return (
    <Pressable
      key={priority}
      onPress={onPress}
      className="flex-row items-center px-4.5 py-4 gap-2"
    >
      <PriorityIcon size={20} priority={priority} />
      <Text className="flex-1">{priority}</Text>
      {isSelected && (
        <SymbolView
          name="checkmark.circle.fill"
          size={20}
          tintColor={colorScheme === "light" ? colors.black : colors.white}
          fallback={
            <Ionicons
              name="checkmark-circle"
              size={20}
              color={colorScheme === "light" ? colors.black : colors.white}
            />
          }
        />
      )}
    </Pressable>
  );
};

export const PriorityBadge = ({
  priority,
  onPriorityChange,
}: {
  priority: StoryPriority;
  onPriorityChange: (priority: StoryPriority) => void;
}) => {
  const priorities: StoryPriority[] = [
    "No Priority",
    "Low",
    "Medium",
    "High",
    "Urgent",
  ];

  return (
    <PropertyBottomSheet
      trigger={
        <Badge color="tertiary">
          <PriorityIcon priority={priority || "No Priority"} />
          <Text>{priority || "No Priority"}</Text>
        </Badge>
      }
      snapPoints={["35%"]}
    >
      <Text className="font-semibold text-center">Priority</Text>
      {priorities.map((p) => (
        <Item
          key={p}
          priority={p}
          onPress={() => onPriorityChange(p)}
          isSelected={priority === p}
        />
      ))}
    </PropertyBottomSheet>
  );
};
