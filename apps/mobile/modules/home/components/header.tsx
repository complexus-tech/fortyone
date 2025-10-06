import React from "react";
import { Pressable } from "react-native";
import { useColorScheme } from "nativewind";
import { SymbolView } from "expo-symbols";
import { Avatar, Row, Text } from "@/components/ui";
import { useRouter } from "expo-router";
import { colors } from "@/constants";
import { useProfile } from "@/modules/users/hooks/use-profile";
import { Host, ContextMenu, HStack, Button, Image } from "@expo/ui/swift-ui";
import { cornerRadius, frame, glassEffect } from "@expo/ui/swift-ui/modifiers";

const getTimeOfDay = () => {
  const hour = new Date().getHours();
  if (hour < 12) return "morning";
  if (hour < 18) return "afternoon";
  return "evening";
};

export const Header = () => {
  const { colorScheme } = useColorScheme();
  const router = useRouter();
  const { data: user } = useProfile();

  return (
    <Row align="center" justify="between" className="mb-4" asContainer>
      <Row align="center" gap={2}>
        <Avatar
          name={user?.fullName || user?.username}
          size="md"
          src={user?.avatarUrl}
        />
        <Text fontSize="2xl" fontWeight="semibold" numberOfLines={1}>
          Good {getTimeOfDay()}!
        </Text>
      </Row>
      <Host matchContents style={{ width: 40, height: 40 }}>
        <ContextMenu>
          <ContextMenu.Items>
            <Button
              systemImage="gear"
              onPress={() => {
                router.push("/settings");
              }}
            >
              Settings
            </Button>
            <Button systemImage="questionmark.circle" onPress={() => {}}>
              Help Center
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
