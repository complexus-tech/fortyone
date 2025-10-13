import { colors } from "@/constants";
import { useTerminology } from "@/hooks";
import { NativeTabs, Icon, Label } from "expo-router/unstable-native-tabs";

export default function TeamLayout() {
  const { getTermDisplay } = useTerminology();
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
      <NativeTabs.Trigger name="sprints" hidden>
        <Icon sf="play.circle" />
        <Label>{sprintsTerm}</Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="objectives" hidden>
        <Icon sf="target" />
        <Label>{objectivesTerm}</Label>
      </NativeTabs.Trigger>
    </NativeTabs>
  );
}
