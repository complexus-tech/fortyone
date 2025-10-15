import React, { useState } from "react";
import { Badge, Text, BottomSheetModal } from "@/components/ui";
import { Story } from "@/modules/stories/types";
import { Pressable } from "react-native";
import { SymbolView } from "expo-symbols";
import { useColorScheme } from "nativewind";
import { colors } from "@/constants";
import { format } from "date-fns";
import { DateTimePicker } from "@expo/ui/swift-ui";

export const EndDateBadge = ({
  story,
  onEndDateChange,
}: {
  story: Story;
  onEndDateChange: (endDate: Date | null) => void;
}) => {
  const { colorScheme } = useColorScheme();
  const [isOpen, setIsOpen] = useState(false);
  const [selectedDate, setSelectedDate] = useState(
    story.endDate ? new Date(story.endDate) : new Date()
  );
  const [hasUserInteracted, setHasUserInteracted] = useState(false);

  const iconColor =
    colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  const handleDateSelected = (date: Date) => {
    setSelectedDate(date);

    // Only call the callback if user has actually interacted
    if (hasUserInteracted) {
      onEndDateChange(date);
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
            {story.endDate
              ? format(new Date(story.endDate), "MMM d")
              : "Add deadline"}
          </Text>
        </Badge>
      </Pressable>

      <BottomSheetModal isOpen={isOpen} spacing={10} onClose={handleClose}>
        <DateTimePicker
          onDateSelected={handleDateSelected}
          displayedComponents="date"
          initialDate={selectedDate.toISOString()}
          variant="graphical"
          color={colorScheme === "dark" ? colors.white : colors.black}
        />
      </BottomSheetModal>
    </>
  );
};
