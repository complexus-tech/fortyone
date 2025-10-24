import React from "react";
import { View, Modal, Pressable } from "react-native";
import { BottomSheetModal as GorhomBottomSheetModal } from "@gorhom/bottom-sheet";

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
    <Modal
      visible={isOpen}
      transparent
      animationType="slide"
      onRequestClose={onClose}
    >
      <View style={{ flex: 1, justifyContent: "flex-end" }}>
        <Pressable
          style={{ flex: 1, backgroundColor: "rgba(0,0,0,0.5)" }}
          onPress={onClose}
        />
        <View
          style={{
            backgroundColor: "white",
            borderTopLeftRadius: 20,
            borderTopRightRadius: 20,
            paddingTop: padding.top,
            paddingHorizontal: padding.leading,
            paddingBottom: padding.bottom,
            gap: spacing,
          }}
        >
          {children}
        </View>
      </View>
    </Modal>
  );
};
