import React, { useState, useMemo } from "react";
import { Badge, Text } from "@/components/ui";
import { Story } from "@/modules/stories/types";
import { Pressable, TextInput } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useColorScheme } from "nativewind";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { cn } from "@/lib/utils";
import { PropertyBottomSheet } from "./property-bottom-sheet";
import { useTeamSprints } from "@/modules/sprints/hooks";

const Item = ({
  sprint,
  onPress,
  isSelected,
  isLast,
}: {
  sprint: { id: string; name: string };
  onPress: () => void;
  isSelected: boolean;
  isLast: boolean;
}) => {
  const { colorScheme } = useColorScheme();
  return (
    <Pressable
      key={sprint.id}
      onPress={onPress}
      className={cn(
        "flex-row items-center p-4 border-b border-gray-100 dark:border-dark-100 gap-2",
        {
          "border-b-0": isLast,
        }
      )}
    >
      <SymbolView
        name="play.circle"
        size={20}
        tintColor={
          colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
        }
        fallback={
          <Ionicons
            name="play-circle"
            size={20}
            color={
              colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
            }
          />
        }
      />
      <Text className="flex-1">{sprint.name}</Text>
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

export const SprintBadge = ({
  story,
  onSprintChange,
}: {
  story: Story;
  onSprintChange: (sprintId: string | null) => void;
}) => {
  const { colorScheme } = useColorScheme();
  const { data: sprints = [] } = useTeamSprints(story.teamId);
  const [searchQuery, setSearchQuery] = useState("");
  const currentSprint = sprints.find((s) => s.id === story.sprintId);
  const iconColor =
    colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  const filteredSprints = useMemo(() => {
    if (!searchQuery.trim()) return sprints;
    return sprints.filter((sprint) =>
      sprint.name.toLowerCase().includes(searchQuery.toLowerCase())
    );
  }, [sprints, searchQuery]);

  return (
    <PropertyBottomSheet
      trigger={
        <Badge color="tertiary">
          <SymbolView
            name="play.circle"
            size={16}
            tintColor={iconColor}
            weight="semibold"
          />
          <Text>{currentSprint?.name || "Add Sprint"}</Text>
        </Badge>
      }
      snapPoints={["25%", "70%"]}
    >
      <Text className="font-semibold mb-2 text-center">Sprint</Text>

      {/* Search Input */}
      <Pressable className="mx-4 mb-4 p-3 bg-gray-50 dark:bg-dark-100 rounded-lg">
        <TextInput
          value={searchQuery}
          onChangeText={setSearchQuery}
          placeholder="Search sprints..."
          placeholderTextColor={colors.gray[300]}
          className="text-base"
          style={{
            color: colors.black,
          }}
        />
      </Pressable>

      {filteredSprints.map((sprint, idx) => (
        <Item
          key={sprint.id}
          sprint={sprint}
          onPress={() => onSprintChange(sprint.id)}
          isSelected={story.sprintId === sprint.id}
          isLast={idx === filteredSprints.length - 1}
        />
      ))}
    </PropertyBottomSheet>
  );
};
