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
            colorScheme === "light" ? colors.white : colors.dark.DEFAULT,
        }}
        handleIndicatorStyle={{
          backgroundColor:
            colorScheme === "light" ? colors.gray[200] : colors.gray[250],
        }}
      >
        <BottomSheetView className="p-5 pb-24">
          <Text
            className="text-lg font-semibold mb-5 text-center"
            style={{
              color: colorScheme === "light" ? colors.black : colors.white,
            }}
          >
            Change Priority
          </Text>
          {priorities.map((priorityOption) => (
            <TouchableOpacity
              key={priorityOption}
              onPress={() => handlePrioritySelect(priorityOption)}
              className="flex-row items-center p-4 border-b border-gray-100 dark:border-dark-100"
            >
              <PriorityIcon priority={priorityOption} />
              <Text
                className="ml-3 text-base flex-1"
                style={{
                  color: colorScheme === "light" ? colors.black : colors.white,
                }}
              >
                {priorityOption}
              </Text>
              {priorityOption === priority && (
                <SymbolView
                  name="checkmark.circle.fill"
                  size={18}
                  tintColor={
                    colorScheme === "light" ? colors.black : colors.white
                  }
                  fallback={
                    <Ionicons
                      name="checkmark-circle"
                      size={18}
                      color={
                        colorScheme === "light" ? colors.black : colors.white
                      }
                    />
                  }
                />
              )}
            </TouchableOpacity>
          ))}
        </BottomSheetView>
      </BottomSheetModal>
    </>
  );
};
