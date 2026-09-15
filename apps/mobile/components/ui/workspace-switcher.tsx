import type { Workspace } from "@/types/workspace";
import {
  Pressable,
  ScrollView,
  StyleSheet,
  View,
  useColorScheme,
  useWindowDimensions,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { Avatar } from "./Avatar";
import { Text } from "./Text";
import { BottomSheetModal } from "./bottom-sheet-modal";
import { themeColors } from "@/constants/colors";
import { useWorkspaces, useCurrentWorkspace } from "@/lib/hooks";
import { useSwitchWorkspace } from "@/lib/hooks/use-switch-workspace";

const WorkspaceItem = ({
  isActive,
  workspace,
  pendingId,
  onSelect,
}: {
  isActive?: boolean;
  workspace: Workspace;
  pendingId: string | null;
  onSelect: (workspace: Workspace) => Promise<void>;
}) => {
  const dark = useColorScheme() === "dark";
  return (
    <Pressable
      onPress={() => {
        if (!pendingId) void onSelect(workspace);
      }}
      disabled={pendingId !== null}
      accessibilityRole="button"
      accessibilityLabel={`${workspace.name}, ${isActive ? "current workspace" : workspace.userRole}`}
      accessibilityState={{
        disabled: pendingId !== null,
        selected: !!isActive,
        busy: pendingId === workspace.id,
      }}
      style={({ pressed }) => [
        styles.row,
        pressed && {
          backgroundColor: themeColors[dark ? "dark" : "light"].surfaceMuted,
        },
      ]}
    >
      <Avatar
        style={{
          width: 40,
          height: 40,
          backgroundColor: workspace.avatarUrl ? undefined : workspace.color,
        }}
        name={workspace.name}
        rounded="xl"
        src={workspace.avatarUrl}
      />
      <View style={styles.details}>
        <Text fontWeight="medium" numberOfLines={1}>
          {workspace.name}
        </Text>
        <Text fontSize="sm" color="muted" numberOfLines={1}>
          {pendingId === workspace.id
            ? "Switching…"
            : `${workspace.slug} · ${workspace.userRole}`}
        </Text>
      </View>
      {isActive ? (
        <Ionicons
          accessible={false}
          name="checkmark"
          size={20}
          color={themeColors[dark ? "dark" : "light"].foreground}
        />
      ) : null}
    </Pressable>
  );
};

export const WorkspaceSwitcher = ({
  isOpened,
  setIsOpened,
}: {
  isOpened: boolean;
  setIsOpened: (isOpened: boolean) => void;
}) => {
  const { height } = useWindowDimensions();
  const { pendingId, selectWorkspace } = useSwitchWorkspace(() =>
    setIsOpened(false),
  );
  const { data: workspaces = [] } = useWorkspaces();
  const { workspace } = useCurrentWorkspace();
  return (
    <BottomSheetModal
      isOpen={isOpened}
      onClose={() => setIsOpened(false)}
      spacing={16}
      padding={{ leading: 20, trailing: 20, top: 20, bottom: 16 }}
    >
      <View style={styles.heading}>
        <Text fontSize="xl" fontWeight="semibold">
          Choose a workspace
        </Text>
        <Text fontSize="sm" color="muted">
          Select the workspace you want to use.
        </Text>
      </View>
      <ScrollView
        style={{ maxHeight: height * 0.6 }}
        contentContainerStyle={styles.list}
        showsVerticalScrollIndicator={false}
      >
        {workspaces.map((item) => (
          <WorkspaceItem
            key={item.id}
            isActive={item.id === workspace?.id}
            workspace={item}
            pendingId={pendingId}
            onSelect={selectWorkspace}
          />
        ))}
      </ScrollView>
    </BottomSheetModal>
  );
};

const styles = StyleSheet.create({
  heading: { gap: 6 },
  list: { gap: 4 },
  row: {
    minHeight: 68,
    flexDirection: "row",
    alignItems: "center",
    paddingVertical: 12,
    gap: 12,
    borderRadius: 10,
  },
  details: { flex: 1, minWidth: 0, gap: 3 },
});
