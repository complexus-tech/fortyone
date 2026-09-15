import type { ComponentProps } from "react";
import { useState } from "react";
import {
  ActivityIndicator,
  Alert,
  Linking,
  Pressable,
  StyleSheet,
  View,
  useColorScheme,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { toast } from "sonner-native";
import { Text, ThemeSwitcher, WorkspaceSwitcher } from "@/components/ui";
import { useProfile } from "@/modules/users/hooks/use-profile";
import { useTheme } from "@/hooks";
import { colors, themeColors } from "@/constants/colors";
import { useAuthStore } from "@/store";
import { useCurrentWorkspace } from "@/lib/hooks/use-workspaces";

type SettingsRowProps = {
  label: string;
  value?: string;
  icon?: ComponentProps<typeof Ionicons>["name"];
  onPress?: () => void;
  destructive?: boolean;
  busy?: boolean;
  external?: boolean;
};

const SettingsRow = ({
  label,
  value,
  icon,
  onPress,
  destructive = false,
  busy = false,
  external = false,
}: SettingsRowProps) => {
  const dark = useColorScheme() === "dark";
  const content = (
    <>
      <Text
        numberOfLines={1}
        ellipsizeMode="tail"
        color={destructive ? "danger" : undefined}
        style={[styles.label, !value && styles.labelOnly]}
      >
        {label}
      </Text>
      {value ? (
        <Text
          color="muted"
          align="right"
          numberOfLines={1}
          ellipsizeMode="tail"
          style={styles.value}
        >
          {value}
        </Text>
      ) : null}
      {busy ? (
        <ActivityIndicator color={colors.danger} size="small" />
      ) : icon ? (
        <Ionicons
          accessible={false}
          name={icon}
          size={18}
          color={
            destructive
              ? colors.danger
              : themeColors[dark ? "dark" : "light"].textMuted
          }
        />
      ) : null}
    </>
  );

  if (!onPress) return <View style={styles.row}>{content}</View>;

  return (
    <Pressable
      accessibilityRole={external ? "link" : "button"}
      accessibilityLabel={value ? `${label}, ${value}` : label}
      accessibilityState={{ busy, disabled: busy }}
      disabled={busy}
      onPress={onPress}
      style={({ pressed }) => [
        styles.row,
        pressed && {
          backgroundColor: themeColors[dark ? "dark" : "light"].surfaceMuted,
        },
      ]}
    >
      {content}
    </Pressable>
  );
};

const GroupDivider = () => {
  const dark = useColorScheme() === "dark";
  return (
    <View
      style={[
        styles.divider,
        { backgroundColor: themeColors[dark ? "dark" : "light"].border },
      ]}
    />
  );
};

export const Form = () => {
  const [isWorkspaceOpen, setIsWorkspaceOpen] = useState(false);
  const [isAppearanceOpen, setIsAppearanceOpen] = useState(false);
  const [isSigningOut, setIsSigningOut] = useState(false);
  const { theme } = useTheme();
  const { workspace } = useCurrentWorkspace();
  const { data: profile } = useProfile();
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const appearance =
    theme === "system" ? "Automatic" : theme === "dark" ? "Dark" : "Light";

  const handleSignOut = () => {
    Alert.alert("Sign Out", "Are you sure you want to sign out?", [
      { text: "Cancel", style: "cancel" },
      {
        text: "Sign Out",
        style: "destructive",
        onPress: () => {
          setIsSigningOut(true);
          void clearAuth()
            .catch((error: unknown) => {
              toast.error("Could not sign out", {
                description:
                  error instanceof Error ? error.message : "Please try again.",
              });
            })
            .finally(() => setIsSigningOut(false));
        },
      },
    ]);
  };

  const handleFeedback = () => {
    void Linking.openURL("https://fortyone.app/contact").catch(() => {
      toast.error("Could not open feedback", {
        description: "Please try again.",
      });
    });
  };

  return (
    <>
      <SettingsRow
        label="Workspace"
        value={workspace?.name || "Choose workspace"}
        icon="chevron-expand"
        onPress={() => setIsWorkspaceOpen(true)}
      />
      <SettingsRow
        label="Appearance"
        value={appearance}
        icon="chevron-expand"
        onPress={() => setIsAppearanceOpen(true)}
      />
      <GroupDivider />
      <Text color="muted" fontSize="xs" style={styles.sectionLabel}>
        Account
      </Text>
      <SettingsRow label="Name" value={profile?.fullName || "—"} />
      <SettingsRow label="Email" value={profile?.email || "—"} />
      <SettingsRow
        label="Username"
        value={profile?.username ? `@${profile.username}` : "—"}
      />
      <GroupDivider />
      <SettingsRow
        label="Send feedback"
        icon="open-outline"
        onPress={handleFeedback}
        external
      />
      <GroupDivider />
      <SettingsRow
        label={isSigningOut ? "Signing out…" : "Log out"}
        icon="log-out-outline"
        onPress={handleSignOut}
        destructive
        busy={isSigningOut}
      />
      <ThemeSwitcher
        isOpened={isAppearanceOpen}
        setIsOpened={setIsAppearanceOpen}
      />
      <WorkspaceSwitcher
        isOpened={isWorkspaceOpen}
        setIsOpened={setIsWorkspaceOpen}
      />
    </>
  );
};

const styles = StyleSheet.create({
  row: {
    minHeight: 56,
    paddingHorizontal: 20,
    paddingVertical: 14,
    flexDirection: "row",
    alignItems: "center",
    gap: 10,
  },
  label: { flexShrink: 1 },
  labelOnly: { flex: 1 },
  value: { flex: 1, minWidth: 0 },
  divider: { height: StyleSheet.hairlineWidth, marginVertical: 8 },
  sectionLabel: { paddingHorizontal: 20, paddingTop: 12, paddingBottom: 4 },
});
