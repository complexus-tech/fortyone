import React, { useState } from "react";
import { Badge, BottomSheetModal, Text as UIText } from "@/components/ui";
import { PriorityIcon } from "@/components/icons";
import { StoryPriority } from "@/modules/stories/types";
import { Pressable } from "react-native";
import { colors } from "@/constants";
import { useTheme } from "@/hooks";
import { HStack, Image, Spacer, Text } from "@expo/ui/swift-ui";
import { frame } from "@expo/ui/swift-ui/modifiers";

const Item = ({
  priority,
  onPress,
  isSelected,
}: {
  priority: StoryPriority;
  onPress: () => void;
  isSelected: boolean;
}) => {
  const { resolvedTheme } = useTheme();
  return (
    <HStack key={priority} onPress={onPress} spacing={6}>
      <HStack modifiers={[frame({ width: 20, height: 20 })]}>
        <PriorityIcon size={20} priority={priority} />
      </HStack>
      <Text size={16}>{priority}</Text>
      <Spacer />
      {isSelected && (
        <Image
          systemName="checkmark.circle.fill"
          size={17}
          color={resolvedTheme === "light" ? colors.black : colors.white}
        />
      )}
    </HStack>
  );
};

export const PriorityBadge = ({
  priority,
  onPriorityChange,
}: {
  priority: StoryPriority;
  onPriorityChange: (priority: StoryPriority) => void;
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const priorities: StoryPriority[] = [
    "No Priority",
    "Low",
    "Medium",
    "High",
    "Urgent",
  ];

  return (
    <>
      <Pressable onPress={() => setIsOpen(true)}>
        <Badge color="tertiary">
          <PriorityIcon priority={priority || "No Priority"} />
          <UIText>{priority || "No Priority"}</UIText>
        </Badge>
      </Pressable>
      <BottomSheetModal
        spacing={24}
        isOpen={isOpen}
        onClose={() => setIsOpen(false)}
        padding={{
          leading: 24,
          trailing: 24,
          top: 36,
          bottom: 5,
        }}
      >
        {priorities.map((p) => (
          <Item
            key={p}
            priority={p}
            onPress={() => {
              onPriorityChange(p);
              setIsOpen(false);
            }}
            isSelected={priority === p}
          />
        ))}
      </BottomSheetModal>
    </>
  );
};
