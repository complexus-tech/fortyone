import { colors } from "@/constants";
import { useFeatures, useSprintsEnabled, useTerminology } from "@/hooks";
import { useLocalSearchParams } from "expo-router";
import { NativeTabs } from "expo-router/unstable-native-tabs";

export default function TeamLayout() {
  const { teamId } = useLocalSearchParams<{ teamId: string }>();
  const { getTermDisplay } = useTerminology();
  const { objectiveEnabled } = useFeatures();
  const sprintsEnabled = useSprintsEnabled(teamId);
  const storyTerm = getTermDisplay("storyTerm", {
    capitalize: true,
    variant: "plural",
  });
  const sprintsTerm = getTermDisplay("sprintTerm", {
    capitalize: true,
    variant: "plural",
  });
  const objectivesTerm = getTermDisplay("objectiveTerm", {
    capitalize: true,
    variant: "plural",
  });
  return (
    <NativeTabs tintColor={colors.primary} minimizeBehavior="onScrollDown">
      <NativeTabs.Trigger name="index">
        <NativeTabs.Trigger.Icon
          sf="rectangle.fill.on.rectangle.angled.fill"
          md="view_list"
        />
        <NativeTabs.Trigger.Label>{storyTerm}</NativeTabs.Trigger.Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="sprints" hidden={!sprintsEnabled}>
        <NativeTabs.Trigger.Icon sf="play.circle" md="play_circle" />
        <NativeTabs.Trigger.Label>{sprintsTerm}</NativeTabs.Trigger.Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="objectives" hidden={!objectiveEnabled}>
        <NativeTabs.Trigger.Icon sf="square.grid.2x2.fill" md="grid_view" />
        <NativeTabs.Trigger.Label>{objectivesTerm}</NativeTabs.Trigger.Label>
      </NativeTabs.Trigger>
    </NativeTabs>
  );
}
