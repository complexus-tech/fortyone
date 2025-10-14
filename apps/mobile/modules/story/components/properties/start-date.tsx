import React, { useState } from "react";
import { Badge, Text, BottomSheetModal } from "@/components/ui";
import { Story } from "@/modules/stories/types";
import { Pressable } from "react-native";
import { SymbolView } from "expo-symbols";
import { useColorScheme } from "nativewind";
import { colors } from "@/constants";
import { format } from "date-fns";
import { DateTimePicker } from "@expo/ui/swift-ui";

export const StartDateBadge = ({
  story,
  onStartDateChange,
}: {
  story: Story;
  onStartDateChange: (startDate: Date | null) => void;
}) => {
  const { colorScheme } = useColorScheme();
  const [isOpen, setIsOpen] = useState(false);
  const [selectedDate, setSelectedDate] = useState(
    story.startDate ? new Date(story.startDate) : new Date()
  );
  const [hasUserInteracted, setHasUserInteracted] = useState(false);

  const iconColor =
    colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  const handleDateSelected = (date: Date) => {
    setSelectedDate(date);

    // Only call the callback if user has actually interacted
    if (hasUserInteracted) {
      onStartDateChange(date);
      setIsOpen(false);
    } else {
      // First time opening, just mark as interacted
      setHasUserInteracted(true);
    }
  };

  const handlePress = () => {
    setHasUserInteracted(false); // Reset interaction flag when opening
    setIsOpen(true);
  };

  const handleClose = () => {
    setIsOpen(false);
  };

  return (
    <>
      <Pressable onPress={handlePress}>
        <Badge color="tertiary">
          <SymbolView name="calendar" size={16} tintColor={iconColor} />
          <Text>
            {story.startDate
              ? format(new Date(story.startDate), "MMM d")
              : "Add start date"}
          </Text>
        </Badge>
      </Pressable>

      <BottomSheetModal isOpen={isOpen} spacing={10} onClose={handleClose}>
        <DateTimePicker
          onDateSelected={handleDateSelected}
          displayedComponents="date"
          initialDate={selectedDate.toISOString()}
          variant="graphical"
          color={colors.primary}
        />
      </BottomSheetModal>
    </>
  );
};
