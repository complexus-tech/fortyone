import type { ContextMenuAction } from "./context-menu.types";
import React from "react";
import { useTheme } from "@/hooks";
import { themeColors } from "@/constants/colors";
import {
  Menu as NativeMenu,
  Host,
  HStack,
  Button,
  Image,
} from "@expo/ui/swift-ui";
import {
  frame,
  glassEffect,
  tint,
  accessibilityLabel,
  accessibilityAddTraits,
} from "@expo/ui/swift-ui/modifiers";
import { StyleProp, ViewStyle } from "react-native";

type ContextMenuButtonProps = {
  actions: ContextMenuAction[];
  children?: React.ReactNode;
  withNoHost?: boolean;
  hostStyle?: StyleProp<ViewStyle>;
  menuLabel?: string;
};

const Menu = ({ actions, children, menuLabel }: ContextMenuButtonProps) => {
  const { resolvedTheme } = useTheme();
  return (
    <NativeMenu
      modifiers={
        menuLabel
          ? [accessibilityLabel(menuLabel)]
          : children
            ? []
            : [accessibilityLabel("Options")]
      }
      label={
        children ?? (
          <HStack
            modifiers={[
              frame({ width: 44, height: 44 }),
              glassEffect({
                glass: {
                  variant: "regular",
                },
              }),
            ]}
          >
            <Image
              systemName="ellipsis"
              size={20}
              color={themeColors[resolvedTheme].foreground}
            />
          </HStack>
        )
      }
    >
      {actions.map((action) => (
        <Button
          key={action.label}
          label={action.label}
          systemImage={action.selected ? "checkmark" : action.systemImage}
          modifiers={[
            ...(action.color ? [tint(action.color)] : []),
            ...(action.selected
              ? [accessibilityAddTraits(["isSelected"])]
              : []),
          ]}
          onPress={action.onPress}
        />
      ))}
    </NativeMenu>
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
    <Host matchContents style={hostStyle}>
      <Menu actions={actions} withNoHost={withNoHost} menuLabel={menuLabel}>
        {children}
      </Menu>
    </Host>
  );
};
