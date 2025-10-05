import React from "react";
import { Row, Text, Back } from "@/components/ui";
import { useGlobalSearchParams } from "expo-router";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { useTerminology } from "@/hooks/use-terminology";
import { colors } from "@/constants";
import { ContextMenu, Host, HStack, Button, Image } from "@expo/ui/swift-ui";
import { cornerRadius, frame, glassEffect } from "@expo/ui/swift-ui/modifiers";
import { useColorScheme } from "nativewind";

export const Header = () => {
  const { colorScheme } = useColorScheme();
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();
  const { data: teams = [] } = useTeams();
  const team = teams.find((team) => team.id === teamId)!;
  const { getTermDisplay } = useTerminology();

  return (
    <Row className="mb-3" asContainer align="center" justify="between">
      <Row align="center" gap={3}>
        <Back />
        <Text fontSize="2xl" fontWeight="semibold">
          {team?.name} /{" "}
          <Text
            fontSize="2xl"
            color="muted"
            fontWeight="semibold"
            className="opacity-80"
          >
            {getTermDisplay("storyTerm", {
              variant: "plural",
              capitalize: true,
            })}
          </Text>
        </Text>
      </Row>
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
