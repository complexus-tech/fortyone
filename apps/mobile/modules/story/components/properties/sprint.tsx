import React, { useState, useMemo } from "react";
import { Badge, Text, Row } from "@/components/ui";
import { Story } from "@/modules/stories/types";
import { Pressable, TextInput } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { truncateText } from "@/lib/utils";
import { PropertyBottomSheet } from "./property-bottom-sheet";
import { useTeamSprints } from "@/modules/sprints/hooks";
import { useTerminology, useTheme } from "@/hooks";

const Item = ({
  sprint,
  onPress,
  isSelected,
}: {
  sprint: { id: string; name: string };
  onPress: () => void;
  isSelected: boolean;
}) => {
  const { resolvedTheme } = useTheme();
  return (
    <Pressable
      key={sprint.id}
      onPress={onPress}
      className="flex-row items-center px-4.5 py-3.5 gap-2"
    >
      <SymbolView
        name="play.circle"
        size={20}
        tintColor={
          resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
        }
        fallback={
          <Ionicons
            name="play-circle"
            size={20}
            color={
              resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
            }
          />
        }
      />
      <Text className="flex-1">{sprint.name}</Text>
      {isSelected && (
        <SymbolView
          name="checkmark.circle.fill"
          size={20}
          tintColor={resolvedTheme === "light" ? colors.black : colors.white}
          fallback={
            <Ionicons
              name="checkmark-circle"
              size={20}
              color={resolvedTheme === "light" ? colors.black : colors.white}
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
  const { getTermDisplay } = useTerminology();
  const { resolvedTheme } = useTheme();
  const { data: sprints = [] } = useTeamSprints(story.teamId);
  const [searchQuery, setSearchQuery] = useState("");
  const currentSprint = sprints.find((s) => s.id === story.sprintId);
  const iconColor =
    resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  const filteredSprints = useMemo(() => {
    let filtered = sprints;

    if (searchQuery.trim()) {
      filtered = filtered.filter((sprint) =>
        sprint.name.toLowerCase().includes(searchQuery.toLowerCase())
      );
    }

    // Limit to maximum 8 sprints
    return filtered.slice(0, 8);
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
          <Text>{truncateText(currentSprint?.name || "Add Sprint", 16)}</Text>
        </Badge>
      }
      snapPoints={["93%"]}
    >
      <Text className="font-semibold mb-3 text-center">Sprint</Text>
      <Row
        className="bg-gray-100/60 dark:bg-dark-100 rounded-xl pl-3 pr-2.5 mx-3.5 mb-1"
        align="center"
        gap={2}
      >
        <SymbolView
          name="magnifyingglass"
          size={20}
          tintColor={
            resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[200]
          }
        />
        <TextInput
          className="flex-1 h-11 font-medium text-[16px] dark:text-white"
          placeholder="Search sprints..."
          placeholderTextColor={
            resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[200]
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
                resolvedTheme === "light"
                  ? colors.gray.DEFAULT
                  : colors.gray[200]
              }
            />
          </Pressable>
        )}
      </Row>

      {filteredSprints.length > 0 ? (
        filteredSprints.map((sprint) => (
          <Item
            key={sprint.id}
            sprint={sprint}
            onPress={() => onSprintChange(sprint.id)}
            isSelected={story.sprintId === sprint.id}
          />
        ))
      ) : sprints.length === 0 ? (
        <Text className="text-center py-8 px-4" color="muted">
          No {getTermDisplay("sprintTerm", { variant: "plural" })} available
        </Text>
      ) : (
        <Text className="text-center py-8 px-4" color="muted">
          No {getTermDisplay("sprintTerm", { variant: "plural" })} found
          matching &quot;{searchQuery}&quot;
        </Text>
      )}
    </PropertyBottomSheet>
  );
};
