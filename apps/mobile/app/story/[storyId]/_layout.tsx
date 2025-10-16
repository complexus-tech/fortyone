import { NativeTabs, Icon, Label } from "expo-router/unstable-native-tabs";
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
        <Icon sf="circle.grid.2x2.fill" />
        <Label>Overview</Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="sub-stories">
        <Icon sf="checklist" />
        <Label>{`Sub ${storyTerm}`}</Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="links">
        <Icon sf="grid" />
        <Label>Links</Label>
      </NativeTabs.Trigger>
      {/* <NativeTabs.Trigger name="attachments">
        <Icon sf="text.document.fill" />
        <Label>Attachments</Label>
      </NativeTabs.Trigger> */}
    </NativeTabs>
  );
}
