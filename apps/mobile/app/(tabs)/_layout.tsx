import { colors } from "@/constants";
import { useUnreadNotifications } from "@/modules/notifications/hooks/use-unread-notifications";
import {
  NativeTabs,
  Icon,
  Label,
  Badge,
} from "expo-router/unstable-native-tabs";

export default function TabsLayout() {
  const { data: unreadNotifications = 0 } = useUnreadNotifications();
  const getBadgeLabel = () => {
    if (unreadNotifications > 9) {
      return "9+";
    } else if (unreadNotifications > 20) {
      return "20+";
    }
    return unreadNotifications.toString();
  };
  return (
    <NativeTabs tintColor={colors.primary} minimizeBehavior="onScrollDown">
      <NativeTabs.Trigger name="index">
        <Icon sf="circle.grid.2x2.fill" />
        <Label>Home</Label>
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="maya">
        <Label>Maya</Label>
        <Icon sf="sparkles" />
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="my-work">
        <Label>My Work</Label>
        <Icon sf="person.fill" />
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="inbox">
        <Label>Inbox</Label>
        <Icon sf="bell.fill" />
        {unreadNotifications > 0 && <Badge>{getBadgeLabel()}</Badge>}
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="search" role="search">
        <Label>Search</Label>
      </NativeTabs.Trigger>
    </NativeTabs>
  );
}
