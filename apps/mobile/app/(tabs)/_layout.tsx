import { colors } from "@/constants";
import { useUnreadNotifications } from "@/modules/notifications/hooks/use-unread-notifications";
import { NativeTabs } from "expo-router/unstable-native-tabs";

export default function TabsLayout() {
  const { data: unreadNotifications = 0 } = useUnreadNotifications();
  const badgeLabel =
    unreadNotifications > 99 ? "99+" : String(unreadNotifications);
  return (
    <NativeTabs tintColor={colors.primary} minimizeBehavior="onScrollDown">
      <NativeTabs.Trigger name="index">
        <NativeTabs.Trigger.Icon sf="circle.grid.2x2.fill" md="dashboard" />
        <NativeTabs.Trigger.Label>Home</NativeTabs.Trigger.Label>
      </NativeTabs.Trigger>

      <NativeTabs.Trigger name="my-work">
        <NativeTabs.Trigger.Label>My Work</NativeTabs.Trigger.Label>
        <NativeTabs.Trigger.Icon sf="person.fill" md="person" />
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="maya" hidden />
      <NativeTabs.Trigger name="inbox">
        <NativeTabs.Trigger.Label>Inbox</NativeTabs.Trigger.Label>
        <NativeTabs.Trigger.Icon sf="bell.fill" md="notifications" />
        {unreadNotifications > 0 && (
          <NativeTabs.Trigger.Badge>{badgeLabel}</NativeTabs.Trigger.Badge>
        )}
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="search" role="search">
        <NativeTabs.Trigger.Label>Search</NativeTabs.Trigger.Label>
        <NativeTabs.Trigger.Icon sf="magnifyingglass" md="search" />
      </NativeTabs.Trigger>
    </NativeTabs>
  );
}
