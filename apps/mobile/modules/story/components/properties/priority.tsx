import React, { useCallback, useRef, useMemo } from "react";
import { Badge, Text } from "@/components/ui";
import { PriorityIcon } from "@/components/icons";
import { StoryPriority } from "@/modules/stories/types";
import { Pressable, StyleSheet, TouchableOpacity } from "react-native";
import { BottomSheetModal, BottomSheetView } from "@gorhom/bottom-sheet";
import { Ionicons } from "@expo/vector-icons";
import { useColorScheme } from "nativewind";

export const PriorityBadge = ({
  priority,
  onPriorityChange,
}: {
  priority: StoryPriority;
  onPriorityChange?: (priority: StoryPriority) => void;
}) => {
  const bottomSheetModalRef = useRef<BottomSheetModal>(null);
  const { colorScheme } = useColorScheme();

  const priorities: StoryPriority[] = [
    "No Priority",
    "Low",
    "Medium",
    "High",
    "Urgent",
  ];

  const snapPoints = useMemo(() => ["25%", "50%"], []);

  // callbacks
  const handlePresentModalPress = useCallback(() => {
    bottomSheetModalRef.current?.present();
  }, []);

  const handleSheetChanges = useCallback((index: number) => {
    console.log("handleSheetChanges", index);
  }, []);

  const handlePrioritySelect = useCallback(
    (selectedPriority: StoryPriority) => {
      onPriorityChange?.(selectedPriority);
      bottomSheetModalRef.current?.dismiss();
    },
    [onPriorityChange]
  );

  const textColor = colorScheme === "light" ? "#000" : "#fff";
  const borderColor = colorScheme === "light" ? "#f0f0f0" : "#333";

  return (
    <>
      <Pressable onPress={handlePresentModalPress}>
        <Badge color="tertiary">
          <PriorityIcon priority={priority || "No Priority"} />
          <Text className="text-[15px]">{priority || "No Priority"}</Text>
        </Badge>
      </Pressable>

      <BottomSheetModal
        ref={bottomSheetModalRef}
        onChange={handleSheetChanges}
        snapPoints={snapPoints}
        index={1}
        backgroundStyle={{
          backgroundColor: colorScheme === "light" ? "#fff" : "#1a1a1a",
        }}
        handleIndicatorStyle={{
          backgroundColor: colorScheme === "light" ? "#ccc" : "#555",
        }}
      >
        <BottomSheetView style={styles.contentContainer}>
          <Text style={[styles.title, { color: textColor }]}>
            Change Priority
          </Text>
          {priorities.map((priorityOption) => (
            <TouchableOpacity
              key={priorityOption}
              onPress={() => handlePrioritySelect(priorityOption)}
              style={[styles.priorityItem, { borderBottomColor: borderColor }]}
            >
              <PriorityIcon priority={priorityOption} />
              <Text style={[styles.priorityText, { color: textColor }]}>
                {priorityOption}
              </Text>
              {priorityOption === priority && (
                <Ionicons
                  name="checkmark"
                  size={20}
                  color={colorScheme === "light" ? "#007AFF" : "#0A84FF"}
                />
              )}
            </TouchableOpacity>
          ))}
        </BottomSheetView>
      </BottomSheetModal>
    </>
  );
};

const styles = StyleSheet.create({
  contentContainer: {
    padding: 20,
    paddingBottom: 100,
  },
  title: {
    fontSize: 18,
    fontWeight: "600",
    marginBottom: 20,
    textAlign: "center",
  },
  priorityItem: {
    flexDirection: "row",
    alignItems: "center",
    padding: 16,
    borderBottomWidth: 1,
  },
  priorityText: {
    marginLeft: 12,
    fontSize: 16,
    flex: 1,
  },
});
