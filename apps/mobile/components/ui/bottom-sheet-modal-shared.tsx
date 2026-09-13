import React from "react";
import { View, Modal, Pressable } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { useTheme } from "@/hooks";
import { colors } from "@/constants";

type BottomSheetModalProps = {
  isOpen: boolean;
  onClose: () => void;
  children: React.ReactNode;
  showDragIndicator?: boolean;
  /** iOS only: children are Expo SwiftUI views rather than React Native views. */
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
  spacing = 20,
  padding = {
    leading: 24,
    trailing: 24,
    top: 32,
    bottom: 5,
  },
}: BottomSheetModalProps) => {
  const { bottom } = useSafeAreaInsets();
  const { resolvedTheme } = useTheme();
  return (
    <Modal
      visible={isOpen}
      transparent
      animationType="slide"
      onRequestClose={onClose}
    >
      <View style={{ flex: 1, justifyContent: "flex-end" }}>
        <Pressable
          accessibilityRole="button"
          accessibilityLabel="Close sheet"
          style={{ flex: 1, backgroundColor: "rgba(0,0,0,0.5)" }}
          onPress={onClose}
        />
        <View
          accessibilityViewIsModal
          style={{
            backgroundColor:
              resolvedTheme === "dark" ? colors.dark[100] : colors.white,
            maxHeight: "85%",
            borderTopLeftRadius: 20,
            borderTopRightRadius: 20,
            paddingTop: padding.top,
            paddingLeft: padding.leading ?? 24,
            paddingRight: padding.trailing ?? 24,
            paddingBottom: Math.max(padding.bottom ?? 16, bottom + 8),
            gap: spacing,
          }}
        >
          {showDragIndicator ? (
            <View
              style={{
                alignSelf: "center",
                width: 36,
                height: 4,
                borderRadius: 2,
                backgroundColor: colors.gray[300],
              }}
            />
          ) : null}
          {children}
        </View>
      </View>
    </Modal>
  );
};
