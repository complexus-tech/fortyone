import React, { useState, useMemo } from "react";
import { Badge, Text } from "@/components/ui";
import { Story } from "@/modules/stories/types";
import { Pressable, TextInput } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useColorScheme } from "nativewind";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { cn, truncateText } from "@/lib/utils";
import { PropertyBottomSheet } from "./property-bottom-sheet";
import { useTeamObjectives } from "@/modules/objectives/hooks/use-objectives";

const Item = ({
  objective,
  onPress,
  isSelected,
  isLast,
}: {
  objective: { id: string; name: string };
  onPress: () => void;
  isSelected: boolean;
  isLast: boolean;
}) => {
  const { colorScheme } = useColorScheme();
  return (
    <Pressable
      key={objective.id}
      onPress={onPress}
      className={cn(
        "flex-row items-center p-4 border-b border-gray-100 dark:border-dark-100 gap-2",
        {
          "border-b-0": isLast,
        }
      )}
    >
      <SymbolView
        name="square.grid.2x2.fill"
        size={20}
        tintColor={
          colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
        }
        fallback={
          <Ionicons
            name="grid"
            size={20}
            color={
              colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
            }
          />
        }
      />
      <Text className="flex-1">{objective.name}</Text>
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

export const ObjectiveBadge = ({
  story,
  onObjectiveChange,
}: {
  story: Story;
  onObjectiveChange: (objectiveId: string | null) => void;
}) => {
  const { colorScheme } = useColorScheme();
  const { data: objectives = [] } = useTeamObjectives(story.teamId);
  const [searchQuery, setSearchQuery] = useState("");
  const currentObjective = objectives.find((o) => o.id === story.objectiveId);
  const iconColor =
    colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  const filteredObjectives = useMemo(() => {
    if (!searchQuery.trim()) return objectives;
    return objectives.filter((objective) =>
      objective.name.toLowerCase().includes(searchQuery.toLowerCase())
    );
  }, [objectives, searchQuery]);

  return (
    <PropertyBottomSheet
      trigger={
        <Badge color="tertiary">
          <SymbolView
            name="square.grid.2x2.fill"
            size={15}
            tintColor={iconColor}
            fallback={<Ionicons name="grid" size={15} color={iconColor} />}
          />
          <Text>
            {truncateText(currentObjective?.name || "Add Objective", 16)}
          </Text>
        </Badge>
      }
      snapPoints={["25%", "70%"]}
    >
      <Text className="font-semibold mb-2 text-center">Objective</Text>

      {/* Search Input */}
      <Pressable className="mx-4 mb-4 p-3 bg-gray-50 dark:bg-dark-100 rounded-lg">
        <TextInput
          value={searchQuery}
          onChangeText={setSearchQuery}
          placeholder="Search objectives..."
          placeholderTextColor={colors.gray[300]}
          className="text-base"
          style={{
            color: colors.black,
          }}
        />
      </Pressable>

      {filteredObjectives.map((objective, idx) => (
        <Item
          key={objective.id}
          objective={objective}
          onPress={() => onObjectiveChange(objective.id)}
          isSelected={story.objectiveId === objective.id}
          isLast={idx === filteredObjectives.length - 1}
        />
      ))}
    </PropertyBottomSheet>
  );
};
