import React, { useCallback, useRef, useMemo } from "react";
import { Badge, Text } from "@/components/ui";
import { PriorityIcon } from "@/components/icons";
import { StoryPriority } from "@/modules/stories/types";
import { Pressable, TouchableOpacity } from "react-native";
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
        <BottomSheetView className="p-5 pb-24">
          <Text className="text-lg font-semibold mb-5 text-center text-gray-900 dark:text-gray-100">
            Change Priority
          </Text>
          {priorities.map((priorityOption) => (
            <TouchableOpacity
              key={priorityOption}
              onPress={() => handlePrioritySelect(priorityOption)}
              className="flex-row items-center p-4 border-b border-gray-200 dark:border-gray-700"
            >
              <PriorityIcon priority={priorityOption} />
              <Text className="ml-3 text-base flex-1 text-gray-900 dark:text-gray-100">
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
