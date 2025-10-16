import React from "react";
import { Badge, Text } from "@/components/ui";
import { Dot } from "@/components/icons";
import { Story } from "@/modules/stories/types";
import { Pressable } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useColorScheme } from "nativewind";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { PropertyBottomSheet } from "./property-bottom-sheet";
import { useTeamStatuses } from "@/modules/statuses/hooks/use-statuses";
import { hexToRgba } from "@/lib/utils/colors";
import { truncateText } from "@/lib/utils";

const Item = ({
  status,
  onPress,
  isSelected,
}: {
  status: { id: string; name: string; color: string };
  onPress: () => void;
  isSelected: boolean;
}) => {
  const { colorScheme } = useColorScheme();
  return (
    <Pressable
      key={status.id}
      onPress={onPress}
      className="flex-row items-center px-4.5 py-4 gap-2"
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
          <Text>{truncateText(currentStatus?.name || "No Status", 16)}</Text>
        </Badge>
      }
      snapPoints={["35%", "70%", "90%"]}
    >
      <Text className="font-semibold mb-3 text-center">Status</Text>
      {statuses.map((status) => (
        <Item
          key={status.id}
          status={status}
          onPress={() => onStatusChange(status.id)}
          isSelected={story.statusId === status.id}
        />
      ))}
    </PropertyBottomSheet>
  );
};
