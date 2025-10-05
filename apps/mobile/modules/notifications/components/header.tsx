import React from "react";
import { Row, Text } from "@/components/ui";
import { colors } from "@/constants";
import { ContextMenu, Host, HStack, Button, Image } from "@expo/ui/swift-ui";
import { cornerRadius, frame, glassEffect } from "@expo/ui/swift-ui/modifiers";
import { useColorScheme } from "nativewind";

export const Header = () => {
  const { colorScheme } = useColorScheme();
  return (
    <Row justify="between" align="center" asContainer className="mb-3">
      <Text fontSize="2xl" fontWeight="semibold" color="black">
        Notifications
      </Text>

      <Host matchContents>
        <ContextMenu>
          <ContextMenu.Items>
            <Button systemImage="gear" onPress={() => {}}>
              Notification settings
            </Button>
            <Button systemImage="checkmark.circle.fill" onPress={() => {}}>
              Mark all as read
            </Button>
            <Button systemImage="delete.forward.fill" onPress={() => {}}>
              Delete read
            </Button>
            <Button systemImage="trash.fill" onPress={() => {}}>
              Delete all
            </Button>
          </ContextMenu.Items>
          <ContextMenu.Trigger>
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
          </ContextMenu.Trigger>
        </ContextMenu>
      </Host>
    </Row>
  );
};
