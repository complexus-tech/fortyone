import React from "react";
import { Host, BottomSheet, Group, VStack } from "@expo/ui/swift-ui";
import {
  foregroundStyle,
  padding as paddingModifier,
  presentationDragIndicator,
} from "@expo/ui/swift-ui/modifiers";
import { BottomSheetModal as ReactNativeSheet } from "./bottom-sheet-modal-shared";
import { useTheme } from "@/hooks";
import { themeColors } from "@/constants/colors";

type BottomSheetModalProps = {
  isOpen: boolean;
  onClose: () => void;
  children: React.ReactNode;
  showDragIndicator?: boolean;
  nativeContent?: boolean;
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
  nativeContent = false,
  spacing = 20,
  padding = {
    leading: 24,
    trailing: 24,
    top: 32,
    bottom: 5,
  },
}: BottomSheetModalProps) => {
  const { resolvedTheme } = useTheme();
  if (!nativeContent) {
    return (
      <ReactNativeSheet
        isOpen={isOpen}
        onClose={onClose}
        showDragIndicator={showDragIndicator}
        spacing={spacing}
        padding={padding}
      >
        {children}
      </ReactNativeSheet>
    );
  }
  return (
    <Host
      matchContents
      colorScheme={resolvedTheme}
      style={{ position: "absolute" }}
    >
      <BottomSheet
        isPresented={isOpen}
        onIsPresentedChange={(presented) => {
          if (!presented) onClose();
        }}
        fitToContents
      >
        <Group
          modifiers={[
            presentationDragIndicator(showDragIndicator ? "visible" : "hidden"),
          ]}
        >
          <VStack
            spacing={spacing}
            modifiers={[
              foregroundStyle(themeColors[resolvedTheme].foreground),
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
        </Group>
      </BottomSheet>
    </Host>
  );
};
