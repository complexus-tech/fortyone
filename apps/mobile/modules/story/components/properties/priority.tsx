import React, { useCallback, useRef, useMemo } from "react";
import { Badge, Text } from "@/components/ui";
import { PriorityIcon } from "@/components/icons";
import { StoryPriority } from "@/modules/stories/types";
import { Pressable, TouchableOpacity } from "react-native";
import { BottomSheetModal, BottomSheetView } from "@gorhom/bottom-sheet";
import { Ionicons } from "@expo/vector-icons";
import { useColorScheme } from "nativewind";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { cn } from "@/lib/utils";

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
        enablePanDownToClose={true}
        enableDismissOnClose={true}
        backdropComponent={({ style }) => (
          <Pressable
            style={[
              style,
              {
                backgroundColor:
                  colorScheme === "light"
                    ? "rgba(0, 0, 0, 0.5)"
                    : "rgba(0, 0, 0, 0.7)",
              },
            ]}
            onPress={() => bottomSheetModalRef.current?.dismiss()}
          />
        )}
        backgroundStyle={{
          backgroundColor:
            colorScheme === "light" ? colors.white : colors.dark[200],
          borderTopLeftRadius: 28,
          borderTopRightRadius: 28,
        }}
        handleIndicatorStyle={{
          backgroundColor:
            colorScheme === "light" ? colors.gray[300] : colors.gray[300],
        }}
      >
        <BottomSheetView className="pb-6">
          <Text className="font-semibold mb-2 text-center">Priority</Text>
          {priorities.map((priorityOption, idx) => (
            <Pressable
              key={priorityOption}
              onPress={() => handlePrioritySelect(priorityOption)}
              className={cn(
                "flex-row items-center p-4 border-b border-gray-100 dark:border-dark-100 gap-2",
                {
                  "border-b-0": idx === priorities.length - 1,
                }
              )}
            >
              <PriorityIcon size={20} priority={priorityOption} />
              <Text
                className="flex-1"
                style={{
                  color: colorScheme === "light" ? colors.black : colors.white,
                }}
              >
                {priorityOption}
              </Text>
              {priorityOption === priority && (
                <SymbolView
                  name="checkmark.circle.fill"
                  size={20}
                  tintColor={
                    colorScheme === "light" ? colors.black : colors.white
                  }
                  fallback={
                    <Ionicons
                      name="checkmark-circle"
                      size={20}
                      color={
                        colorScheme === "light" ? colors.black : colors.white
                      }
                    />
                  }
                />
              )}
            </Pressable>
          ))}
        </BottomSheetView>
      </BottomSheetModal>
    </>
  );
};
