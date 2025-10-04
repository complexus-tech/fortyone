import { Text, Row, Back } from "@/components/ui";
import { Pressable } from "react-native";
import { colors } from "@/constants";
import { SymbolView } from "expo-symbols";
import { useGlobalSearchParams } from "expo-router";
import { useStory } from "@/modules/stories/hooks";
import { useTeams } from "@/modules/teams/hooks/use-teams";

export const Header = () => {
  const { storyId } = useGlobalSearchParams<{ storyId: string }>();
  const { data: story } = useStory(storyId);
  const { data: teams = [] } = useTeams();
  const team = teams.find((team) => team.id === story?.teamId);
  return (
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
      >
        <SymbolView name="ellipsis" tintColor={colors.dark[50]} />
      </Pressable>
    </Row>
  );
};
