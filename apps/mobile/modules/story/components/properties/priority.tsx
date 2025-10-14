import React from "react";
import { Badge, Text } from "@/components/ui";
import { PriorityIcon } from "@/components/icons";
import { StoryPriority } from "@/modules/stories/types";
import { Pressable } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useColorScheme } from "nativewind";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { cn } from "@/lib/utils";
import { PropertyBottomSheet } from "./property-bottom-sheet";

const Item = ({
  priority,
  onPress,
  isSelected,
  isLast,
}: {
  priority: StoryPriority;
  onPress: () => void;
  isSelected: boolean;
  isLast: boolean;
}) => {
  const { colorScheme } = useColorScheme();
  return (
    <Pressable
      key={priority}
      onPress={onPress}
      className={cn(
        "flex-row items-center p-4 border-b border-gray-100 dark:border-dark-100 gap-2",
        {
          "border-b-0": isLast,
        }
      )}
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
  onPriorityChange?: (priority: StoryPriority) => void;
}) => {
  const priorities: StoryPriority[] = [
    "No Priority",
    "Low",
    "Medium",
    "High",
    "Urgent",
  ];

  const handlePrioritySelect = (priority: StoryPriority) => {
    onPriorityChange?.(priority);
  };

  return (
    <PropertyBottomSheet
      title="Priority"
      trigger={
        <Badge color="tertiary">
          <PriorityIcon priority={priority || "No Priority"} />
          <Text className="text-[15px]">{priority || "No Priority"}</Text>
        </Badge>
      }
      snapPoints={["25%", "50%"]}
    >
      {priorities.map((p, idx) => (
        <Item
          key={p}
          priority={p}
          onPress={() => handlePrioritySelect(p)}
          isSelected={priority === p}
          isLast={idx === priorities.length - 1}
        />
      ))}
    </PropertyBottomSheet>
  );
};
