import React, { useState } from "react";
import { useTheme } from "@/hooks";
import { colors } from "@/constants";
import {
  Pressable,
  ScrollView,
  StyleProp,
  Text,
  View,
  ViewStyle,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { BottomSheetModal } from "./bottom-sheet-modal";

type ContextMenuAction = {
  systemImage?: string;
  label: string;
  onPress: () => void;
  color?: string;
};

type ContextMenuButtonProps = {
  actions: ContextMenuAction[];
  children?: React.ReactNode;
  withNoHost?: boolean;
  hostStyle?: StyleProp<ViewStyle>;
};

const Menu = ({ actions, children }: ContextMenuButtonProps) => {
  const { resolvedTheme } = useTheme();
  const [isOpen, setIsOpen] = useState(false);
  const foreground =
    resolvedTheme === "light" ? colors.dark[50] : colors.gray[200];
  return (
    <>
      <Pressable
        accessibilityRole="button"
        accessibilityLabel="More options"
        accessibilityState={{
          expanded: isOpen,
          disabled: actions.length === 0,
        }}
        disabled={actions.length === 0}
        hitSlop={6}
        style={{
          width: 40,
          height: 40,
          borderRadius: 20,
          backgroundColor:
            resolvedTheme === "light"
              ? "rgba(255, 255, 255, 0.8)"
              : "rgba(0, 0, 0, 0.8)",
          justifyContent: "center",
          alignItems: "center",
          shadowColor: "#000",
          shadowOffset: {
            width: 0,
            height: 2,
          },
          shadowOpacity: 0.1,
          shadowRadius: 4,
          elevation: 3,
        }}
        onPress={() => setIsOpen(true)}
      >
        {children ? (
          children
        ) : (
          <Ionicons name="ellipsis-horizontal" size={20} color={foreground} />
        )}
      </Pressable>
      <BottomSheetModal
        isOpen={isOpen}
        onClose={() => setIsOpen(false)}
        spacing={8}
      >
        <Text
          accessibilityRole="header"
          style={{ color: foreground, fontSize: 18, fontWeight: "600" }}
        >
          Options
        </Text>
        <ScrollView keyboardShouldPersistTaps="handled">
          {actions.map((action) => (
            <Pressable
              key={action.label}
              accessibilityRole="button"
              accessibilityLabel={action.label}
              onPress={() => {
                setIsOpen(false);
                action.onPress();
              }}
              style={({ pressed }) => ({
                minHeight: 48,
                paddingVertical: 14,
                opacity: pressed ? 0.6 : 1,
              })}
            >
              <Text style={{ fontSize: 16, color: action.color ?? foreground }}>
                {action.label}
              </Text>
            </Pressable>
          ))}
        </ScrollView>
        <Pressable
          accessibilityRole="button"
          onPress={() => setIsOpen(false)}
          style={{
            minHeight: 44,
            justifyContent: "center",
            alignItems: "center",
          }}
        >
          <Text style={{ fontSize: 16, color: foreground }}>Cancel</Text>
        </Pressable>
      </BottomSheetModal>
    </>
  );
};

export const ContextMenuButton = ({
  actions,
  children,
  withNoHost,
  hostStyle = {
    width: 40,
    height: 40,
  },
}: ContextMenuButtonProps) => {
  if (withNoHost) {
    return (
      <Menu actions={actions} withNoHost={withNoHost}>
        {children}
      </Menu>
    );
  }
  return (
    <View style={hostStyle}>
      <Menu actions={actions} withNoHost={withNoHost}>
        {children}
      </Menu>
    </View>
  );
};
