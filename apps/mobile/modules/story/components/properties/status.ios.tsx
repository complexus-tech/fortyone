import React, { useState } from "react";
import { Badge, BottomSheetModal, Text as UIText } from "@/components/ui";
import { Dot } from "@/components/icons";
import { Story } from "@/modules/stories/types";
import { Pressable } from "react-native";
import { colors } from "@/constants";
import { useTheme } from "@/hooks";
import { HStack, Image, Spacer, Text } from "@expo/ui/swift-ui";
import { frame } from "@expo/ui/swift-ui/modifiers";
import { useTeamStatuses } from "@/modules/statuses/hooks/use-statuses";
import { hexToRgba } from "@/lib/utils/colors";
import { truncateText } from "@/lib/utils";

const Item = ({
  status,
  onPress,
  isSelected,
}: {
  status: { id: string; name: string; color: string };
  onPress: () => void;
  isSelected: boolean;
}) => {
  const { resolvedTheme } = useTheme();
  return (
    <HStack key={status.id} onPress={onPress} spacing={6}>
      <HStack modifiers={[frame({ width: 12, height: 12 })]}>
        <Dot color={status.color} size={4} />
      </HStack>
      <Text size={16}>{status.name}</Text>
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

export const StatusBadge = ({
  story,
  onStatusChange,
}: {
  story: Story;
  onStatusChange: (statusId: string) => void;
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const { data: statuses = [] } = useTeamStatuses(story.teamId);
  const currentStatus = statuses.find((s) => s.id === story.statusId);

  return (
    <>
      <Pressable onPress={() => setIsOpen(true)}>
        <Badge
          style={{
            backgroundColor: hexToRgba(currentStatus?.color, 0.1),
            borderColor: hexToRgba(currentStatus?.color, 0.2),
          }}
        >
          <Dot color={currentStatus?.color} size={12} />
          <UIText>
            {truncateText(currentStatus?.name || "No Status", 16)}
          </UIText>
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
        {statuses.map((status) => (
          <Item
            key={status.id}
            status={status}
            onPress={() => {
              onStatusChange(status.id);
              setIsOpen(false);
            }}
            isSelected={story.statusId === status.id}
          />
        ))}
      </BottomSheetModal>
    </>
  );
};
