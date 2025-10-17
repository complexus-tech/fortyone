import React from "react";
import { useTheme } from "@/hooks";
import { colors } from "@/constants";
import { ContextMenu, Host, HStack, Button, Image } from "@expo/ui/swift-ui";
import { frame, glassEffect } from "@expo/ui/swift-ui/modifiers";
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
  const { resolvedTheme } = useTheme();
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
