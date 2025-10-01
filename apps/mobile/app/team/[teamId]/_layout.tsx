import { colors } from "@/constants";
import { NativeTabs, Icon, Label } from "expo-router/unstable-native-tabs";

export default function TeamLayout() {
  return (
    <NativeTabs tintColor={colors.primary} minimizeBehavior="onScrollDown">
      <NativeTabs.Trigger name="stories">
        <Icon sf="rectangle.fill.on.rectangle.angled.fill" />
        <Label>Stories</Label>
      </NativeTabs.Trigger>
      {/* <NativeTabs.Trigger name="backlog">
        <Icon sf="circle.dashed" />
        <Label>Backlog</Label>
      </NativeTabs.Trigger> */}
      <NativeTabs.Trigger name="sprints">
        <Icon sf="play.circle" />
        <Label>Sprints</Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="objectives">
        <Icon sf="target" />
        <Label>Objectives</Label>
      </NativeTabs.Trigger>
      {/* add search tab */}
      {/* <NativeTabs.Trigger name="objectives">
        <Icon sf="magnifyingglass" />
        <Label>Search</Label>
      </NativeTabs.Trigger> */}
    </NativeTabs>
  );
}
