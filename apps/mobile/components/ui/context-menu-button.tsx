import type { ContextMenuAction } from "./context-menu.types";
import React, { useState } from "react";
import { useTheme } from "@/hooks";
import { themeColors } from "@/constants/colors";
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

type ContextMenuButtonProps = {
  actions: ContextMenuAction[];
  children?: React.ReactNode;
  withNoHost?: boolean;
  hostStyle?: StyleProp<ViewStyle>;
  menuLabel?: string;
};

const Menu = ({
  actions,
  children,
  menuLabel = "More options",
}: ContextMenuButtonProps) => {
  const { resolvedTheme } = useTheme();
  const [isOpen, setIsOpen] = useState(false);
  const theme = themeColors[resolvedTheme];
  const foreground = theme.foreground;
  return (
    <>
      <Pressable
        accessibilityRole="button"
        accessibilityLabel={menuLabel}
        accessibilityState={{
          expanded: isOpen,
          disabled: actions.length === 0,
        }}
        disabled={actions.length === 0}
        hitSlop={6}
        style={({ pressed }) => ({
          width: 44,
          height: 44,
          borderRadius: 12,
          backgroundColor: pressed ? theme.stateActive : "transparent",
          justifyContent: "center",
          alignItems: "center",
        })}
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
          style={{ color: foreground, fontSize: 18, fontWeight: "700" }}
        >
          Options
        </Text>
        <ScrollView keyboardShouldPersistTaps="handled">
          {actions.map((action) => (
            <Pressable
              key={action.label}
              accessibilityRole="button"
              accessibilityLabel={action.label}
              accessibilityState={{ selected: action.selected }}
              onPress={() => {
                setIsOpen(false);
                action.onPress();
              }}
              style={({ pressed }) => ({
                minHeight: 48,
                paddingVertical: 14,
                flexDirection: "row",
                alignItems: "center",
                gap: 12,
                opacity: pressed ? 0.6 : 1,
              })}
            >
              <Text
                style={{
                  flex: 1,
                  fontSize: 16,
                  color: action.color ?? foreground,
                }}
              >
                {action.label}
              </Text>
              {action.selected ? (
                <Ionicons
                  accessible={false}
                  name="checkmark"
                  size={20}
                  color={foreground}
                />
              ) : null}
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
  menuLabel,
  hostStyle = {
    width: 44,
    height: 44,
  },
}: ContextMenuButtonProps) => {
  if (withNoHost) {
    return (
      <Menu actions={actions} withNoHost={withNoHost} menuLabel={menuLabel}>
        {children}
      </Menu>
    );
  }
  return (
    <View style={hostStyle}>
      <Menu actions={actions} withNoHost={withNoHost} menuLabel={menuLabel}>
        {children}
      </Menu>
    </View>
  );
};
