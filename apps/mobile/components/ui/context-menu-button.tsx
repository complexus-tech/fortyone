import React from "react";
import { useColorScheme } from "nativewind";
import { colors } from "@/constants";
import { ContextMenu, Host, HStack, Button, Image } from "@expo/ui/swift-ui";
import { cornerRadius, frame, glassEffect } from "@expo/ui/swift-ui/modifiers";
import { SFSymbol } from "expo-symbols";

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
};

const Menu = ({ actions, children }: ContextMenuButtonProps) => {
  const { colorScheme } = useColorScheme();
  return (
    <ContextMenu>
      <ContextMenu.Items>
        {actions.map((action, index) => (
          <Button
            key={index}
            systemImage={action.systemImage}
            color={action.color}
            onPress={action.onPress}
          >
            {action.label}
          </Button>
        ))}
      </ContextMenu.Items>
      <ContextMenu.Trigger>
        {children ? (
          children
        ) : (
          <HStack
            modifiers={[
              frame({ width: 40, height: 40 }),
              glassEffect({
                glass: {
                  variant: "regular",
                },
              }),
              cornerRadius(18),
            ]}
          >
            <Image
              systemName="ellipsis"
              size={20}
              color={
                colorScheme === "light" ? colors.dark[50] : colors.gray[300]
              }
            />
          </HStack>
        )}
      </ContextMenu.Trigger>
    </ContextMenu>
  );
};

export const ContextMenuButton = ({
  actions,
  children,
  withNoHost,
}: ContextMenuButtonProps) => {
  if (withNoHost) {
    return (
      <Menu actions={actions} withNoHost={withNoHost}>
        {children}
      </Menu>
    );
  }
  return (
    <Host matchContents style={{ width: 40, height: 40 }}>
      <Menu actions={actions} withNoHost={withNoHost}>
        {children}
      </Menu>
    </Host>
  );
};
