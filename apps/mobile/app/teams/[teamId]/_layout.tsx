import { colors } from "@/constants";
import { useFeatures, useSprintsEnabled, useTerminology } from "@/hooks";
import { useGlobalSearchParams } from "expo-router";
import { NativeTabs, Icon, Label } from "expo-router/unstable-native-tabs";

export default function TeamLayout() {
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();
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
        <Icon sf="rectangle.fill.on.rectangle.angled.fill" />
        <Label>{storyTerm}</Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="sprints" hidden={!sprintsEnabled}>
        <Icon sf="play.circle" />
        <Label>{sprintsTerm}</Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="objectives" hidden={!objectiveEnabled}>
        <Icon sf="square.grid.2x2.fill" />
        <Label>{objectivesTerm}</Label>
      </NativeTabs.Trigger>
    </NativeTabs>
  );
}
