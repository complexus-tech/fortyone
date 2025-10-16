import React, { useState, useMemo } from "react";
import { Badge, Text, Row } from "@/components/ui";
import { Story } from "@/modules/stories/types";
import { Pressable, TextInput } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useColorScheme } from "nativewind";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { PropertyBottomSheet } from "./property-bottom-sheet";
import { Dot } from "@/components/icons";
import { useLabels } from "@/modules/labels/hooks/use-labels";
import { useCreateLabelMutation } from "@/modules/labels/hooks/use-create-label-mutation";
import type { Label } from "@/types";

const Item = ({
  label,
  onPress,
  isSelected,
}: {
  label: Label;
  onPress: () => void;
  isSelected: boolean;
}) => {
  const { colorScheme } = useColorScheme();
  return (
    <Pressable
      key={label.id}
      onPress={onPress}
      className="flex-row items-center px-4.5 py-4 gap-2"
    >
      <Dot color={label.color} size={12} />
      <Text className="flex-1">{label.name}</Text>
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

export const LabelsBadge = ({
  story,
  onLabelsChange,
}: {
  story: Story;
  onLabelsChange: (labelIds: string[]) => void;
}) => {
  const { colorScheme } = useColorScheme();
  const [searchQuery, setSearchQuery] = useState("");
  const { data: labels = [] } = useLabels();
  const { mutateAsync: createLabel, isPending } = useCreateLabelMutation();
  const iconColor =
    colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  // Filter labels based on search query and limit to 8
  const filteredLabels = useMemo(() => {
    let filtered = labels as Label[];

    if (searchQuery.trim()) {
      const lowerCaseQuery = searchQuery.toLowerCase();
      filtered = filtered.filter((label: Label) =>
        label.name.toLowerCase().includes(lowerCaseQuery)
      );
    }

    // Limit to maximum 8 labels
    return filtered.slice(0, 8);
  }, [labels, searchQuery]);

  const handleLabelToggle = (labelId: string) => {
    const currentLabelIds = story.labels || [];
    const isSelected = currentLabelIds.includes(labelId);

    if (isSelected) {
      onLabelsChange(currentLabelIds.filter((id) => id !== labelId));
    } else {
      onLabelsChange([...currentLabelIds, labelId]);
    }
  };

  const handleCreateLabel = () => {
    createLabel({
      name: searchQuery,
      color: "red",
      teamId: story.teamId,
    }).then((response) => {
      if (response.data) {
        // Auto-select the new label
        onLabelsChange([...(story.labels || []), response.data.id]);
        // Don't clear search - keep the input so user can see the created label
      }
    });
  };

  return (
    <PropertyBottomSheet
      trigger={
        <Badge color="tertiary">
          <SymbolView name="tag.fill" size={16} tintColor={iconColor} />
          <Text>
            {story.labels?.length
              ? `${story.labels.length} labels`
              : "Add labels"}
          </Text>
        </Badge>
      }
      snapPoints={["93%"]}
    >
      <Text className="font-semibold mb-3 text-center">Labels</Text>
      <Row
        className="bg-gray-100/60 dark:bg-dark-50 rounded-xl pl-3 pr-2.5 mx-3.5"
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
          placeholder="Search labels..."
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

      {filteredLabels.length === 0 && searchQuery && (
        <Pressable
          onPress={handleCreateLabel}
          disabled={isPending}
          className="flex-row items-center p-4 gap-2"
        >
          <SymbolView
            name="plus.circle.fill"
            size={20}
            tintColor={isPending ? colors.gray[300] : colors.gray.DEFAULT}
          />
          <Text className="flex-1">
            {isPending ? "Creating..." : `Create new label: "${searchQuery}"`}
          </Text>
        </Pressable>
      )}

      {filteredLabels.map((label) => (
        <Item
          key={label.id}
          label={label}
          onPress={() => handleLabelToggle(label.id)}
          isSelected={story.labels?.includes(label.id) || false}
        />
      ))}
    </PropertyBottomSheet>
  );
};
