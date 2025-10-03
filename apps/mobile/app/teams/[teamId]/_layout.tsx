import { colors } from "@/constants";
import { useLocalSearchParams } from "expo-router";
import { NativeTabs, Icon, Label } from "expo-router/unstable-native-tabs";

export default function TeamLayout() {
  const { teamId } = useLocalSearchParams<{ teamId: string }>();
  return (
    <NativeTabs tintColor={colors.primary} minimizeBehavior="onScrollDown">
      <NativeTabs.Trigger name="index">
        <Icon sf="rectangle.fill.on.rectangle.angled.fill" />
        <Label>Stories</Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="sprints">
        <Icon sf="play.circle" />
        <Label>Sprints</Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="objectives">
        <Icon sf="target" />
        <Label>Objectives</Label>
      </NativeTabs.Trigger>
    </NativeTabs>
  );
}
