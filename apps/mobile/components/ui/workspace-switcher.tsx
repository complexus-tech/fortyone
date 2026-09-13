import React from "react";
import { Avatar } from "./avatar";
import { BottomSheetModal } from "./bottom-sheet-modal";
import { colors } from "@/constants";
import { Pressable, View, Text as RNText } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { Workspace } from "@/types/workspace";
import { useWorkspaces, useCurrentWorkspace } from "@/lib/hooks";
import { useTheme } from "@/hooks";
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
  const handleSwitchWorkspace = () => {
    if (!pendingId) void onSelect(workspace);
  };

  const { resolvedTheme } = useTheme();
  const mutedTextColor =
    resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300];
  return (
    <Pressable
      onPress={handleSwitchWorkspace}
      disabled={pendingId !== null}
      accessibilityRole="button"
      accessibilityState={{
        disabled: pendingId !== null,
        selected: !!isActive,
        busy: pendingId === workspace.id,
      }}
      style={{
        flexDirection: "row",
        alignItems: "center",
        paddingVertical: 12,
        paddingHorizontal: 16,
        gap: 8,
      }}
    >
      <View style={{ width: 38, height: 38 }}>
        <Avatar
          style={{
            backgroundColor: workspace.avatarUrl ? undefined : workspace.color,
          }}
          name={workspace.name}
          className="size-[38px]"
          rounded="xl"
          src={workspace.avatarUrl}
        />
      </View>
      <View style={{ flex: 1 }}>
        <RNText
          style={{
            fontSize: 16,
            fontWeight: "500",
            color: resolvedTheme === "light" ? "black" : "white",
          }}
        >
          {workspace.name}
        </RNText>
        <RNText
          style={{
            fontSize: 14,
            fontWeight: "500",
            color: mutedTextColor,
          }}
        >
          {pendingId === workspace.id ? "Switching…" : workspace.userRole}
        </RNText>
      </View>
      {isActive && (
        <Ionicons
          name="checkmark-circle"
          size={20}
          color={
            resolvedTheme === "light" ? colors.dark.DEFAULT : colors.gray[200]
          }
        />
      )}
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
  const { pendingId, selectWorkspace } = useSwitchWorkspace(() =>
    setIsOpened(false),
  );
  const { resolvedTheme } = useTheme();
  const { data: workspaces = [] } = useWorkspaces();
  const { workspace } = useCurrentWorkspace();
  const mutedTextColor =
    resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300];
  return (
    <BottomSheetModal
      isOpen={isOpened}
      onClose={() => setIsOpened(false)}
      spacing={18}
    >
      <View style={{ paddingHorizontal: 16, paddingBottom: 16 }}>
        <RNText
          style={{
            fontWeight: "600",
            color: mutedTextColor,
            fontSize: 14,
            marginBottom: 8,
          }}
        >
          Switch Workspace
        </RNText>
        {workspaces.map((wk) => (
          <WorkspaceItem
            key={wk.id}
            isActive={wk.id === workspace?.id}
            workspace={wk}
            pendingId={pendingId}
            onSelect={selectWorkspace}
          />
        ))}
      </View>
    </BottomSheetModal>
  );
};
