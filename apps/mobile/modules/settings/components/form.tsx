import type { ComponentProps } from "react";
import { useRef, useState } from "react";
import {
  ActivityIndicator,
  Alert,
  Pressable,
  StyleSheet,
  View,
  useColorScheme,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { openBrowserAsync } from "expo-web-browser";
import { toast } from "sonner-native";
import { Text, ThemeSwitcher, WorkspaceSwitcher } from "@/components/ui";
import { useProfile } from "@/modules/users/hooks/use-profile";
import { getProfile } from "@/modules/users/queries/get-profile";
import { useTheme } from "@/hooks";
import { colors, themeColors } from "@/constants/colors";
import { useAuthStore } from "@/store";
import { useCurrentWorkspace } from "@/lib/hooks/use-workspaces";
import { openPublicPage } from "@/lib/public-pages";
import { HttpError } from "@/lib/http";
import { getApplicationURL } from "@/lib/http/config";
import {
  AccountDeletionCleanupError,
  type AccountDeletionScope,
} from "@/lib/account-deletion";
import { useDeleteAccountMutation } from "../hooks/use-delete-account-mutation";
import { isAccountDeletionSession } from "../actions/delete-account";

type SettingsRowProps = {
  label: string;
  value?: string;
  icon?: ComponentProps<typeof Ionicons>["name"];
  onPress?: () => void;
  destructive?: boolean;
  busy?: boolean;
  disabled?: boolean;
  external?: boolean;
};

const SettingsRow = ({
  label,
  value,
  icon,
  onPress,
  destructive = false,
  busy = false,
  disabled = false,
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
      accessibilityState={{ busy, disabled: busy || disabled }}
      disabled={busy || disabled}
      onPress={onPress}
      style={({ pressed }) => [
        styles.row,
        disabled && { opacity: 0.5 },
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
  const [deletionError, setDeletionError] = useState<string | null>(null);
  const [deletionAccepted, setDeletionAccepted] = useState(false);
  const [needsOwnershipResolution, setNeedsOwnershipResolution] =
    useState(false);
  const [isOpeningResolution, setIsOpeningResolution] = useState(false);
  const deletionStage = useRef<"confirming" | "deleting" | null>(null);
  const deleteMutation = useDeleteAccountMutation();
  const { theme } = useTheme();
  const { workspace } = useCurrentWorkspace();
  const { data: profile } = useProfile();
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const appearance =
    theme === "system" ? "Automatic" : theme === "dark" ? "Dark" : "Light";
  const isBusy = isSigningOut || deleteMutation.isPending;

  const runDeletion = (scope: AccountDeletionScope) => {
    if (!isAccountDeletionSession(scope)) {
      deletionStage.current = null;
      return;
    }
    deletionStage.current = "deleting";
    setDeletionError(null);
    setNeedsOwnershipResolution(false);
    void deleteMutation
      .mutateAsync(scope)
      .catch((error: unknown) => {
        if (!isAccountDeletionSession(scope)) return;
        if (error instanceof AccountDeletionCleanupError)
          setDeletionAccepted(true);
        setNeedsOwnershipResolution(
          error instanceof HttpError && error.status === 409,
        );
        setDeletionError(
          error instanceof Error
            ? error.message
            : "Account deletion could not be confirmed. Please try again.",
        );
      })
      .finally(() => {
        deletionStage.current = null;
      });
  };

  const handleResolveOwnership = async () => {
    if (isOpeningResolution || isBusy) return;
    const session = useAuthStore.getState();
    if (!session.userId || !session.isAuthenticated || session.isLoading)
      return;
    const scope = {
      userId: session.userId,
      sessionEpoch: session.sessionEpoch,
    };
    setIsOpeningResolution(true);
    // Start with a resolved promise so URL/configuration errors also release
    // the busy state. The browser receives no native credential or account ID.
    const opened = await Promise.resolve()
      .then(() =>
        openBrowserAsync(
          new URL("/auth/account-deletion", getApplicationURL()).toString(),
        ),
      )
      .then(
        () => true,
        () => {
          toast.error("Could not open account settings", {
            description: "Please try again when you are connected.",
          });
          return false;
        },
      )
      .finally(() => setIsOpeningResolution(false));
    if (!opened || !isAccountDeletionSession(scope)) return;
    try {
      // Safari can complete deletion without changing the native AppState.
      // The shared client checks the session epoch before sending, and a 401
      // clears only the matching credential. Do not require a workspace here:
      // the user may have deleted their last one and still need this screen.
      await getProfile();
    } catch {
      if (isAccountDeletionSession(scope)) {
        toast.error("Could not refresh account status", {
          description: "Reconnect and reopen Settings to check your account.",
        });
      }
    }
  };

  const handleDeleteAccount = () => {
    if (isBusy || deletionStage.current) return;
    const session = useAuthStore.getState();
    if (!session.userId || !session.isAuthenticated || session.isLoading)
      return;
    const scope = {
      userId: session.userId,
      sessionEpoch: session.sessionEpoch,
    };
    if (deletionAccepted) {
      runDeletion(scope);
      return;
    }
    deletionStage.current = "confirming";
    const cancelConfirmation = () => {
      if (deletionStage.current === "confirming") deletionStage.current = null;
    };
    Alert.alert(
      "Permanently delete account?",
      "Your profile will be deleted. This cannot be undone.",
      [
        { text: "Cancel", style: "cancel", onPress: cancelConfirmation },
        {
          text: "Delete account",
          style: "destructive",
          onPress: () => runDeletion(scope),
        },
      ],
      { cancelable: true, onDismiss: cancelConfirmation },
    );
  };

  const handleSignOut = () => {
    if (isBusy || deletionStage.current) return;
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

  return (
    <>
      <SettingsRow
        label="Workspace"
        value={workspace?.name || "Choose workspace"}
        icon="chevron-expand"
        onPress={() => setIsWorkspaceOpen(true)}
        disabled={isBusy}
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
        label="Support and feedback"
        icon="open-outline"
        onPress={() => openPublicPage("support")}
        external
      />
      <SettingsRow
        label="Privacy policy"
        icon="open-outline"
        onPress={() => openPublicPage("privacy")}
        external
      />
      <SettingsRow
        label="Terms of service"
        icon="open-outline"
        onPress={() => openPublicPage("terms")}
        external
      />
      <GroupDivider />
      <SettingsRow
        label={isSigningOut ? "Signing out…" : "Log out"}
        icon="log-out-outline"
        onPress={handleSignOut}
        destructive
        busy={isSigningOut}
        disabled={deleteMutation.isPending}
      />
      <SettingsRow
        label={
          deletionAccepted
            ? deleteMutation.isPending
              ? "Signing out…"
              : "Finish signing out"
            : deleteMutation.isPending
              ? "Deleting account…"
              : "Delete account"
        }
        icon="trash-outline"
        onPress={handleDeleteAccount}
        destructive
        busy={deleteMutation.isPending}
        disabled={isSigningOut}
      />
      {deletionError ? (
        <Text
          color="danger"
          accessibilityRole="alert"
          accessibilityLiveRegion="polite"
          style={styles.deletionError}
        >
          {deletionError}
        </Text>
      ) : null}
      {needsOwnershipResolution ? (
        <SettingsRow
          label="Resolve workspace ownership"
          icon="open-outline"
          onPress={() => void handleResolveOwnership()}
          busy={isOpeningResolution}
          disabled={isBusy}
          external
        />
      ) : null}
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
  deletionError: { paddingHorizontal: 20, paddingBottom: 16 },
});
