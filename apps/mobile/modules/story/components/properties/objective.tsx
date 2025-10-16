import React, { useState, useMemo } from "react";
import { Badge, Text, Row } from "@/components/ui";
import { Story } from "@/modules/stories/types";
import { Pressable, TextInput } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useColorScheme } from "nativewind";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { truncateText } from "@/lib/utils";
import { PropertyBottomSheet } from "./property-bottom-sheet";
import { useTeamObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useTerminology } from "@/hooks";

const Item = ({
  objective,
  onPress,
  isSelected,
}: {
  objective: { id: string; name: string };
  onPress: () => void;
  isSelected: boolean;
}) => {
  const { colorScheme } = useColorScheme();
  return (
    <Pressable
      key={objective.id}
      onPress={onPress}
      className="flex-row items-center px-4.5 py-3.5 gap-2"
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
  const { getTermDisplay } = useTerminology();
  const { colorScheme } = useColorScheme();
  const { data: objectives = [] } = useTeamObjectives(story.teamId);
  const [searchQuery, setSearchQuery] = useState("");
  const currentObjective = objectives.find((o) => o.id === story.objectiveId);
  const iconColor =
    colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  const filteredObjectives = useMemo(() => {
    let filtered = objectives;

    if (searchQuery.trim()) {
      filtered = filtered.filter((objective) =>
        objective.name.toLowerCase().includes(searchQuery.toLowerCase())
      );
    }

    // Limit to maximum 8 objectives
    return filtered.slice(0, 8);
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
      snapPoints={["93%"]}
    >
      <Text className="font-semibold mb-3 text-center">Objective</Text>
      <Row
        className="bg-gray-100/60 dark:bg-dark-100 rounded-xl pl-3 pr-2.5 mx-3.5 mb-1"
        align="center"
        gap={2}
      >
        <SymbolView
          name="magnifyingglass"
          size={20}
          tintColor={
            colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[200]
          }
        />
        <TextInput
          className="flex-1 h-11 font-medium text-[16px] dark:text-white"
          placeholder="Search objectives..."
          placeholderTextColor={
            colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[200]
          }
          value={searchQuery}
          onChangeText={setSearchQuery}
          autoFocus
        />
        {searchQuery.length > 0 && (
          <Pressable
            onPress={() => {
              setSearchQuery("");
            }}
            className="p-1"
          >
            <SymbolView
              name="xmark.circle.fill"
              size={20}
              tintColor={
                colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[200]
              }
            />
          </Pressable>
        )}
      </Row>

      {filteredObjectives.length > 0 ? (
        filteredObjectives.map((objective) => (
          <Item
            key={objective.id}
            objective={objective}
            onPress={() => onObjectiveChange(objective.id)}
            isSelected={story.objectiveId === objective.id}
          />
        ))
      ) : objectives.length === 0 ? (
        <Text className="text-center py-8 px-4" color="muted">
          No {getTermDisplay("objectiveTerm", { variant: "plural" })} available
        </Text>
      ) : (
        <Text className="text-center py-8 px-4" color="muted">
          No {getTermDisplay("objectiveTerm", { variant: "plural" })} found
          matching &quot;{searchQuery}&quot;
        </Text>
      )}
    </PropertyBottomSheet>
  );
};
