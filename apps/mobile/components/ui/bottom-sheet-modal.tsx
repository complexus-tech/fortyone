import React from "react";
import { Host, BottomSheet, VStack } from "@expo/ui/swift-ui";
import { padding } from "@expo/ui/swift-ui/modifiers";

type BottomSheetModalProps = {
  isOpen: boolean;
  onClose: () => void;
  children: React.ReactNode;
  showDragIndicator?: boolean;
};

export const BottomSheetModal = ({
  isOpen,
  onClose,
  children,
  showDragIndicator = true,
}: BottomSheetModalProps) => {
  return (
    <Host matchContents style={{ position: "absolute" }}>
      <BottomSheet
        isOpened={isOpen}
        onIsOpenedChange={onClose}
        presentationDragIndicator={showDragIndicator ? "visible" : "hidden"}
      >
        <VStack
          spacing={20}
          modifiers={[padding({ all: 24 })]}
          alignment="leading"
        >
          {children}
        </VStack>
      </BottomSheet>
    </Host>
  );
};
