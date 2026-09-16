import { useState } from "react";
import { Pressable, View } from "react-native";
import { WebIcon } from "@/components/icons/web-icon";
import { useRouter } from "expo-router";
import { Avatar, Text, WorkspaceSwitcher } from "@/components/ui";
import { useProfile } from "@/modules/users/hooks/use-profile";
import { useCurrentWorkspace } from "@/lib/hooks/use-workspaces";
import { useTheme } from "@/hooks";
import { themeColors } from "@/constants/colors";
import { NewStoryButton } from "./new-story";
import { GlassIconButton } from "@/components/ui/glass-icon-button";

export const Header = () => {
  const router = useRouter();
  const { data: profile } = useProfile();
  const { workspace } = useCurrentWorkspace();
  const { resolvedTheme } = useTheme();
  const [switcherOpen, setSwitcherOpen] = useState(false);
  const muted = themeColors[resolvedTheme].textMuted;

  return (
    <>
      <View
        style={{
          flexDirection: "row",
          alignItems: "center",
          paddingHorizontal: 20,
          minHeight: 64,
          gap: 8,
        }}
      >
        <Pressable
          accessibilityRole="button"
          accessibilityLabel={`Switch workspace${workspace?.name ? `, ${workspace.name}` : ""}`}
          onPress={() => setSwitcherOpen(true)}
          style={{
            flex: 1,
            minWidth: 0,
            minHeight: 44,
            flexDirection: "row",
            alignItems: "center",
            gap: 8,
          }}
        >
          <Avatar
            name={workspace?.name}
            src={workspace?.avatarUrl}
            rounded="md"
            style={{
              width: 26,
              height: 26,
              backgroundColor: workspace?.avatarUrl
                ? undefined
                : workspace?.color,
            }}
          />
          <Text
            numberOfLines={1}
            style={{
              flexShrink: 1,
              fontSize: 18,
              lineHeight: 24,
              fontWeight: "600",
            }}
          >
            {workspace?.name || "Workspace"}
          </Text>
          <WebIcon name="chevronDown" size={14} color={muted} />
        </Pressable>
        <NewStoryButton />
        <GlassIconButton
          label="Open settings"
          onPress={() => router.push("/settings")}
        >
          <Avatar
            name={profile?.fullName || profile?.username}
            src={profile?.avatarUrl}
            style={{ width: 30, height: 30 }}
          />
        </GlassIconButton>
      </View>
      <WorkspaceSwitcher
        isOpened={switcherOpen}
        setIsOpened={setSwitcherOpen}
      />
    </>
  );
};
