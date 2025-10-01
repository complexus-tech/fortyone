import { colors } from "@/constants";
import { NativeTabs, Icon, Label } from "expo-router/unstable-native-tabs";

export default function StoryLayout() {
  return (
    <NativeTabs tintColor={colors.primary} minimizeBehavior="onScrollDown">
      <NativeTabs.Trigger name="index">
        <Icon sf="doc.text.fill" />
        <Label>Overview</Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="sub-stories">
        <Icon sf="list.bullet" />
        <Label>Sub Stories</Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="links">
        <Icon sf="link" />
        <Label>Links</Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="attachments">
        <Icon sf="paperclip" />
        <Label>Attachments</Label>
      </NativeTabs.Trigger>
    </NativeTabs>
  );
}
