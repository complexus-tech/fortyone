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
  onEndDateChange: (endDate: string | null) => void;
}) => {
  const { colorScheme } = useColorScheme();
  const [isOpen, setIsOpen] = useState(false);
  const [selectedDate, setSelectedDate] = useState(
    story.endDate ? new Date(story.endDate) : new Date()
  );

  const iconColor =
    colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  const handleDateSelected = (date: Date) => {
    setSelectedDate(date);
    console.log(date);
    // onEndDateChange(date.toISOString());
    // setIsOpen(false);
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
          <Text color={story.endDate ? undefined : "muted"}>
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
          color={colors.primary}
        />
      </BottomSheetModal>
    </>
  );
};
