import React from "react";
import { BottomSheetModal } from "@/components/ui";
import { colors } from "@/constants";
import { HStack, Text } from "@expo/ui/swift-ui";

export const Sheet = ({
  isOpened,
  setIsOpened,
}: {
  isOpened: boolean;
  setIsOpened: (isOpened: boolean) => void;
}) => {
  return (
    <BottomSheetModal isOpen={isOpened} onClose={() => setIsOpened(false)}>
      <HStack>
        <Text weight="semibold" color={colors.gray.DEFAULT} size={15}>
          Appearance
        </Text>
      </HStack>
    </BottomSheetModal>
  );
};
