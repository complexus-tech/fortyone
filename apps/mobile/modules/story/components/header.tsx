import { Text, Row, Back } from "@/components/ui";
import { Pressable, useWindowDimensions, View } from "react-native";
import { colors } from "@/constants";
import { SymbolView } from "expo-symbols";
import { useGlobalSearchParams } from "expo-router";
import { useStory } from "@/modules/stories/hooks";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import {
  BottomSheet,
  Host,
  Text as SwiftUIText,
  ContextMenu,
  HStack,
  Button,
  Image,
} from "@expo/ui/swift-ui";
import { useState } from "react";
import {
  padding,
  cornerRadius,
  frame,
  glassEffect,
} from "@expo/ui/swift-ui/modifiers";
import { useColorScheme } from "nativewind";

export const Header = () => {
  const { colorScheme } = useColorScheme();
  const { storyId } = useGlobalSearchParams<{ storyId: string }>();
  const { data: story } = useStory(storyId);
  const { data: teams = [] } = useTeams();
  const team = teams.find((team) => team.id === story?.teamId);
  const { width } = useWindowDimensions();
  const [isOpened, setIsOpened] = useState(false);

  return (
    <>
      <Row className="mb-2" justify="between" align="center" asContainer>
        <Row align="center" gap={3}>
          <Back />
          <Text fontSize="2xl" fontWeight="semibold">
            {team?.code}-
            <Text fontSize="2xl" fontWeight="semibold" color="muted">
              {story?.sequenceId}
            </Text>
          </Text>
        </Row>
        <Host matchContents style={{ width: 40, height: 40 }}>
          <ContextMenu>
            <ContextMenu.Items>
              <Button systemImage="pencil" onPress={() => {}}>
                Edit
              </Button>
              <Button systemImage="archivebox.fill" onPress={() => {}}>
                Archive
              </Button>
              <Button systemImage="link" onPress={() => {}}>
                Copy link
              </Button>
              <Button
                color={colors.danger}
                systemImage="trash.fill"
                onPress={() => {}}
              >
                Delete forever
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
      <Host style={{ position: "absolute", width }}>
        <BottomSheet
          isOpened={isOpened}
          modifiers={[padding({ top: 20, bottom: 20 })]}
          onIsOpenedChange={(e) => setIsOpened(e)}
        >
          <View>
            <Text>Hello, world! again</Text>
          </View>
          <SwiftUIText>Hello, world!</SwiftUIText>
        </BottomSheet>
      </Host>
    </>
  );
};
