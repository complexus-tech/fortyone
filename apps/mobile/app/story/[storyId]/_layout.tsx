import { NativeTabs } from "expo-router/unstable-native-tabs";
import { colors } from "@/constants";
import { useTerminology } from "@/hooks";

export default function StoryLayout() {
  const { getTermDisplay } = useTerminology();
  const storyTerm = getTermDisplay("storyTerm", {
    capitalize: true,
    variant: "plural",
  });
  return (
    <NativeTabs tintColor={colors.primary} minimizeBehavior="onScrollDown">
      <NativeTabs.Trigger name="index">
        <NativeTabs.Trigger.Icon sf="circle.grid.2x2.fill" md="dashboard" />
        <NativeTabs.Trigger.Label>Overview</NativeTabs.Trigger.Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="sub-stories">
        <NativeTabs.Trigger.Icon sf="checklist" md="checklist" />
        <NativeTabs.Trigger.Label>{`Sub ${storyTerm}`}</NativeTabs.Trigger.Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="links">
        <NativeTabs.Trigger.Icon sf="grid" md="link" />
        <NativeTabs.Trigger.Label>Links</NativeTabs.Trigger.Label>
      </NativeTabs.Trigger>
    </NativeTabs>
  );
}
