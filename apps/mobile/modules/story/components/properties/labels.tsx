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
import { Dot } from "@/components/icons";

// Mock data - replace with actual hooks when available
const mockLabels = [
  { id: "1", name: "Bug", color: "#ef4444" },
  { id: "2", name: "Feature", color: "#10b981" },
  { id: "3", name: "Enhancement", color: "#3b82f6" },
  { id: "4", name: "Documentation", color: "#8b5cf6" },
  { id: "5", name: "Design", color: "#f59e0b" },
];

const Item = ({
  label,
  onPress,
  isSelected,
  isLast,
}: {
  label: { id: string; name: string; color: string };
  onPress: () => void;
  isSelected: boolean;
  isLast: boolean;
}) => {
  const { colorScheme } = useColorScheme();
  return (
    <Pressable
      key={label.id}
      onPress={onPress}
      className={cn(
        "flex-row items-center p-4 border-b border-gray-100 dark:border-dark-100 gap-2",
        {
          "border-b-0": isLast,
        }
      )}
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

  // Filter labels based on search query
  const filteredLabels = useMemo(() => {
    if (!searchQuery) {
      return mockLabels;
    }
    const lowerCaseQuery = searchQuery.toLowerCase();
    return mockLabels.filter((label) =>
      label.name.toLowerCase().includes(lowerCaseQuery)
    );
  }, [searchQuery]);

  // Get current labels
  const currentLabels = mockLabels.filter((label) =>
    story.labels?.includes(label.id)
  );

  const handleLabelToggle = (labelId: string) => {
    const currentLabelIds = story.labels || [];
    const isSelected = currentLabelIds.includes(labelId);

    if (isSelected) {
      // Remove label
      onLabelsChange(currentLabelIds.filter((id) => id !== labelId));
    } else {
      // Add label
      onLabelsChange([...currentLabelIds, labelId]);
    }
  };

  const handleCreateLabel = () => {
    // TODO: Implement create label functionality
    console.log("Create label:", searchQuery);
  };

  return (
    <PropertyBottomSheet
      trigger={
        <Badge color="tertiary">
          <SymbolView
            name="tag.fill"
            size={16}
            tintColor={colors.gray.DEFAULT}
          />
          <Text color={story.labels?.length ? undefined : "muted"}>
            {story.labels?.length
              ? `${story.labels.length} labels`
              : "Add labels"}
          </Text>
        </Badge>
      }
      snapPoints={["25%", "70%"]}
    >
      <TextInput
        placeholder="Search labels..."
        value={searchQuery}
        onChangeText={setSearchQuery}
        className="p-3 mb-4 rounded-lg border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-gray-100"
        placeholderTextColor={
          colorScheme === "light" ? colors.gray[300] : colors.gray[200]
        }
      />

      {filteredLabels.length === 0 && searchQuery && (
        <Pressable
          onPress={handleCreateLabel}
          className="flex-row items-center p-4 border-b border-gray-100 dark:border-dark-100 gap-2"
        >
          <SymbolView
            name="plus.circle.fill"
            size={20}
            tintColor={colors.gray.DEFAULT}
          />
          <Text className="flex-1">Create new label: "{searchQuery}"</Text>
        </Pressable>
      )}

      {filteredLabels.map((label, idx) => (
        <Item
          key={label.id}
          label={label}
          onPress={() => handleLabelToggle(label.id)}
          isSelected={story.labels?.includes(label.id) || false}
          isLast={idx === filteredLabels.length - 1}
        />
      ))}
    </PropertyBottomSheet>
  );
};
