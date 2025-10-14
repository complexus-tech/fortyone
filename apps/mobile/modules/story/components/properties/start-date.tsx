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
  onStartDateChange: (startDate: string | null) => void;
}) => {
  const { colorScheme } = useColorScheme();
  const [isOpen, setIsOpen] = useState(false);
  const [selectedDate, setSelectedDate] = useState(
    story.startDate ? new Date(story.startDate) : new Date()
  );

  const iconColor =
    colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  const handleDateSelected = (date: Date) => {
    setSelectedDate(date);
    onStartDateChange(date.toISOString());
    setIsOpen(false);
  };

  const handlePress = () => {
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
          <Text color={story.startDate ? undefined : "muted"}>
            {story.startDate
              ? format(new Date(story.startDate), "MMM d")
              : "Add start date"}
          </Text>
        </Badge>
      </Pressable>

      <BottomSheetModal isOpen={isOpen} spacing={10} onClose={handleClose}>
        <DateTimePicker
          // onDateSelected={handleDateSelected}
          displayedComponents="date"
          initialDate={selectedDate.toISOString()}
          variant="graphical"
          color={colors.primary}
        />
      </BottomSheetModal>
    </>
  );
};
