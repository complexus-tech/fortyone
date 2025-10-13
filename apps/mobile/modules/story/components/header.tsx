import { Text, Row, Back, ContextMenuButton } from "@/components/ui";
import { useGlobalSearchParams } from "expo-router";
import { useStory } from "@/modules/stories/hooks";
import { useTeams } from "@/modules/teams/hooks/use-teams";

export const Header = () => {
  const { storyId } = useGlobalSearchParams<{ storyId: string }>();
  const { data: story } = useStory(storyId);
  const { data: teams = [] } = useTeams();
  const team = teams.find((team) => team.id === story?.teamId);

  return (
    <Row className="mb-3" justify="between" align="center" asContainer>
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        {team?.code}-
        <Text fontSize="2xl" fontWeight="semibold" color="muted">
          {story?.sequenceId}
        </Text>
      </Text>
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
  );
};
