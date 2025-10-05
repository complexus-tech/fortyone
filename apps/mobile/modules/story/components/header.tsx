import { Text, Row, Back } from "@/components/ui";
import { Pressable, useWindowDimensions, View } from "react-native";
import { colors } from "@/constants";
import { SymbolView } from "expo-symbols";
import { useGlobalSearchParams } from "expo-router";
import { useStory } from "@/modules/stories/hooks";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { BottomSheet, Host, Text as SwiftUIText } from "@expo/ui/swift-ui";
import { useState } from "react";
import { padding } from "@expo/ui/swift-ui/modifiers";

export const Header = () => {
  const { storyId } = useGlobalSearchParams<{ storyId: string }>();
  const { data: story } = useStory(storyId);
  const { data: teams = [] } = useTeams();
  const team = teams.find((team) => team.id === story?.teamId);
  const { width } = useWindowDimensions();
  const [isOpened, setIsOpened] = useState(false);

  return (
    <>
      <Row className="pb-3" justify="between" align="center" asContainer>
        <Back />
        <Text fontSize="2xl" fontWeight="semibold">
          {team?.code}-{story?.sequenceId}
        </Text>
        <Pressable
          className="p-2 rounded-md"
          style={({ pressed }) => [
            pressed && { backgroundColor: colors.gray[50] },
          ]}
          onPress={() => setIsOpened(true)}
        >
          <SymbolView name="ellipsis" tintColor={colors.dark[50]} />
        </Pressable>
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
