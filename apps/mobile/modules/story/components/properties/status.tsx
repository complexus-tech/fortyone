import React from "react";
import { Badge, Text } from "@/components/ui";
import { Dot } from "@/components/icons";
import { Story } from "@/modules/stories/types";
import { Pressable } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useColorScheme } from "nativewind";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { cn } from "@/lib/utils";
import { PropertyBottomSheet } from "./property-bottom-sheet";
import { useTeamStatuses } from "@/modules/statuses/hooks/use-statuses";
import { hexToRgba } from "@/lib/utils/colors";

const Item = ({
  status,
  onPress,
  isSelected,
  isLast,
}: {
  status: { id: string; name: string; color: string };
  onPress: () => void;
  isSelected: boolean;
  isLast: boolean;
}) => {
  const { colorScheme } = useColorScheme();
  return (
    <Pressable
      key={status.id}
      onPress={onPress}
      className={cn(
        "flex-row items-center p-4 border-b border-gray-100 dark:border-dark-100 gap-2",
        {
          "border-b-0": isLast,
        }
      )}
    >
      <Dot color={status.color} size={12} />
      <Text className="flex-1">{status.name}</Text>
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

export const StatusBadge = ({
  story,
  onStatusChange,
}: {
  story: Story;
  onStatusChange: (statusId: string) => void;
}) => {
  const { data: statuses = [] } = useTeamStatuses(story.teamId);
  const currentStatus = statuses.find((s) => s.id === story.statusId);

  return (
    <PropertyBottomSheet
      trigger={
        <Badge
          style={{
            backgroundColor: hexToRgba(currentStatus?.color, 0.1),
            borderColor: hexToRgba(currentStatus?.color, 0.2),
          }}
        >
          <Dot color={currentStatus?.color} size={12} />
          <Text>{currentStatus?.name || "No Status"}</Text>
        </Badge>
      }
      snapPoints={["25%", "60%"]}
    >
      <Text className="font-semibold mb-2 text-center">Status</Text>
      {statuses.map((status, idx) => (
        <Item
          key={status.id}
          status={status}
          onPress={() => onStatusChange(status.id)}
          isSelected={story.statusId === status.id}
          isLast={idx === statuses.length - 1}
        />
      ))}
    </PropertyBottomSheet>
  );
};
