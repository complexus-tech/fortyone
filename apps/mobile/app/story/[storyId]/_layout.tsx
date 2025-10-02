import { NativeTabs, Icon, Label } from "expo-router/unstable-native-tabs";
import { colors } from "@/constants";

export default function StoryLayout() {
  return (
    <NativeTabs tintColor={colors.primary} minimizeBehavior="onScrollDown">
      <NativeTabs.Trigger name="index">
        <Icon sf="circle.grid.2x2.fill" />
        <Label>Overview</Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="sub-stories">
        <Icon sf="checklist" />
        <Label>Sub Stories</Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="links">
        <Icon sf="list.number" />
        <Label>Links</Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="attachments">
        <Icon sf="square.fill.text.grid.1x2" />
        <Label>Attachments</Label>
      </NativeTabs.Trigger>
    </NativeTabs>
  );
}
