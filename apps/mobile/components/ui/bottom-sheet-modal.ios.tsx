import React from "react";
import { Host, BottomSheet, VStack } from "@expo/ui/swift-ui";
import { padding as paddingModifier } from "@expo/ui/swift-ui/modifiers";

type BottomSheetModalProps = {
  isOpen: boolean;
  onClose: () => void;
  children: React.ReactNode;
  showDragIndicator?: boolean;
  spacing?: number;
  padding?: {
    leading?: number;
    trailing?: number;
    top?: number;
    bottom?: number;
  };
};

export const BottomSheetModal = ({
  isOpen,
  onClose,
  children,
  showDragIndicator = true,
  spacing = 20,
  padding = {
    leading: 24,
    trailing: 24,
    top: 32,
    bottom: 5,
  },
}: BottomSheetModalProps) => {
  return (
    <Host matchContents style={{ position: "absolute" }}>
      <BottomSheet
        isOpened={isOpen}
        onIsOpenedChange={onClose}
        presentationDragIndicator={showDragIndicator ? "visible" : "hidden"}
      >
        <VStack
          spacing={spacing}
          modifiers={[
            paddingModifier({
              leading: padding.leading,
              trailing: padding.trailing,
              top: padding.top,
              bottom: padding.bottom,
            }),
          ]}
          alignment="leading"
        >
          {children}
        </VStack>
      </BottomSheet>
    </Host>
  );
};
