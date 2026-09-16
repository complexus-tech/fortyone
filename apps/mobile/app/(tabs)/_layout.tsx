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
          src={{
            default: require("@/assets/icons/tabs/home.png"),
            selected: require("@/assets/icons/tabs/home-selected.png"),
          }}
          renderingMode="template"
        />
        <NativeTabs.Trigger.Label hidden>Home</NativeTabs.Trigger.Label>
      </NativeTabs.Trigger>

      <NativeTabs.Trigger name="my-work" accessibilityLabel="My Work">
        <NativeTabs.Trigger.Label hidden>My Work</NativeTabs.Trigger.Label>
        <NativeTabs.Trigger.Icon
          src={{
            default: require("@/assets/icons/tabs/my-work.png"),
            selected: require("@/assets/icons/tabs/my-work-selected.png"),
          }}
          renderingMode="template"
        />
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="inbox" accessibilityLabel="Inbox">
        <NativeTabs.Trigger.Label hidden>Inbox</NativeTabs.Trigger.Label>
        <NativeTabs.Trigger.Icon
          src={{
            default: require("@/assets/icons/tabs/inbox.png"),
            selected: require("@/assets/icons/tabs/inbox-selected.png"),
          }}
          renderingMode="template"
        />
        {unreadNotifications > 0 && (
          <NativeTabs.Trigger.Badge>{badgeLabel}</NativeTabs.Trigger.Badge>
        )}
      </NativeTabs.Trigger>
      <NativeTabs.Trigger
        name="search"
        role="search"
        accessibilityLabel="Search"
      >
        <NativeTabs.Trigger.Label hidden>Search</NativeTabs.Trigger.Label>
        <NativeTabs.Trigger.Icon
          src={require("@/assets/icons/tabs/search.png")}
          renderingMode="template"
        />
      </NativeTabs.Trigger>
    </NativeTabs>
  );
}
