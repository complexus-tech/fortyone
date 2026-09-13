import React from "react";
import { useTheme } from "@/hooks";
import { colors } from "@/constants";
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
} from "@expo/ui/swift-ui/modifiers";
import { SFSymbol } from "expo-symbols";
import { StyleProp, ViewStyle } from "react-native";

type ContextMenuAction = {
  systemImage?: SFSymbol;
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
  return (
    <NativeMenu
      modifiers={children ? [] : [accessibilityLabel("Options")]}
      label={
        children ?? (
          <HStack
            modifiers={[
              frame({ width: 40, height: 40 }),
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
              color={
                resolvedTheme === "light" ? colors.dark[50] : colors.gray[200]
              }
            />
          </HStack>
        )
      }
    >
      {actions.map((action) => (
        <Button
          key={action.label}
          label={action.label}
          systemImage={action.systemImage}
          modifiers={action.color ? [tint(action.color)] : []}
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
    <Host matchContents style={hostStyle}>
      <Menu actions={actions} withNoHost={withNoHost}>
        {children}
      </Menu>
    </Host>
  );
};
