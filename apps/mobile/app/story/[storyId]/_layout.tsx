import { colors } from "@/constants";
import { NativeTabs, Icon, Label } from "expo-router/unstable-native-tabs";

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
        <Icon sf="globe" />
        <Label>Links</Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="attachments">
        <Icon
          sf="books.vertical.fill"
          // sf="book.pages.fill"
        />
        <Label>Attachments</Label>
      </NativeTabs.Trigger>
    </NativeTabs>
  );
}
