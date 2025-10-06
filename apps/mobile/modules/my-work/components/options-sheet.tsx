import React from "react";
import { BottomSheetModal } from "@/components/ui";
import { Text } from "@expo/ui/swift-ui";

export const OptionsSheet = ({
  isOpened,
  setIsOpened,
}: {
  isOpened: boolean;
  setIsOpened: (isOpened: boolean) => void;
}) => {
  return (
    <BottomSheetModal isOpen={isOpened} onClose={() => setIsOpened(false)}>
      <Text>Hello, world!</Text>
    </BottomSheetModal>
  );
};
