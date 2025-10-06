import React from "react";
import { BottomSheetModal } from "@/components/ui";
import { Text, HStack } from "@expo/ui/swift-ui";
import { colors } from "@/constants";

export const OptionsSheet = ({
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
      <Text>Hello, world!</Text>
    </BottomSheetModal>
  );
};
