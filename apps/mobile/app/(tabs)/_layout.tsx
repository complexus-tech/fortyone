import {
  NativeTabs,
  Icon,
  Label,
  Badge,
} from "expo-router/unstable-native-tabs";

export default function TabsLayout() {
  return (
    <NativeTabs tintColor="#f43f5e" minimizeBehavior="onScrollDown">
      <NativeTabs.Trigger name="index">
        <Icon sf="square.grid.2x2.fill" />
        <Label>Home</Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="inbox">
        <Label>Notifications</Label>
        <Icon sf="bell.fill" />
        <Badge>9+</Badge>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="my-work">
        <Label>My Work</Label>
        <Icon sf="person.fill" />
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="settings">
        <Label>AI Assistant</Label>
        <Icon sf="apple.intelligence" />
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="search" role="search">
        <Label>Search</Label>
      </NativeTabs.Trigger>
    </NativeTabs>
  );
}
