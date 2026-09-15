import { themeColors } from "@/constants/colors";
import { useUnreadNotifications } from "@/modules/notifications/hooks/use-unread-notifications";
import { NativeTabs } from "expo-router/unstable-native-tabs";
import { useColorScheme } from "react-native";

export default function TabsLayout() {
  const { data: unreadNotifications = 0 } = useUnreadNotifications();
  const badgeLabel =
    unreadNotifications > 99 ? "99+" : String(unreadNotifications);
  const foreground =
    themeColors[useColorScheme() === "dark" ? "dark" : "light"].foreground;
  return (
    <NativeTabs
      tintColor={foreground}
      iconColor={{ default: foreground, selected: foreground }}
      minimizeBehavior="onScrollDown"
    >
      <NativeTabs.Trigger name="index" accessibilityLabel="Home">
        <NativeTabs.Trigger.Icon
          sf={{
            default: "square.grid.2x2",
            selected: "square.grid.2x2.fill",
          }}
          md="grid_view"
        />
        <NativeTabs.Trigger.Label>Home</NativeTabs.Trigger.Label>
      </NativeTabs.Trigger>

      <NativeTabs.Trigger name="my-work" accessibilityLabel="My Work">
        <NativeTabs.Trigger.Label>My Work</NativeTabs.Trigger.Label>
        <NativeTabs.Trigger.Icon sf="checklist" md="checklist" />
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="maya" hidden />
      <NativeTabs.Trigger name="inbox" accessibilityLabel="Inbox">
        <NativeTabs.Trigger.Label>Inbox</NativeTabs.Trigger.Label>
        <NativeTabs.Trigger.Icon sf="tray.fill" md="inbox" />
        {unreadNotifications > 0 && (
          <NativeTabs.Trigger.Badge>{badgeLabel}</NativeTabs.Trigger.Badge>
        )}
      </NativeTabs.Trigger>
      <NativeTabs.Trigger
        name="search"
        role="search"
        accessibilityLabel="Search"
      >
        <NativeTabs.Trigger.Label>Search</NativeTabs.Trigger.Label>
        <NativeTabs.Trigger.Icon sf="magnifyingglass" md="search" />
      </NativeTabs.Trigger>
    </NativeTabs>
  );
}
