import { Text, Row, Back, ContextMenuButton } from "@/components/ui";
import { useWindowDimensions, View } from "react-native";
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
        <ContextMenuButton
          actions={[
            {
              systemImage: "pencil",
              label: "Edit",
              onPress: () => {},
            },
            {
              systemImage: "archivebox.fill",
              label: "Archive",
              onPress: () => {},
            },
            {
              systemImage: "link",
              label: "Copy link",
              onPress: () => {},
            },
            {
              systemImage: "trash.fill",
              label: "Delete",
              onPress: () => {},
            },
          ]}
        />
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
