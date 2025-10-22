import React from "react";
import { Avatar } from "./avatar";
import * as Updates from "expo-updates";
import { BottomSheetModal } from "./bottom-sheet-modal";
import { colors } from "@/constants";
import { HStack, Image, Spacer, Text, VStack } from "@expo/ui/swift-ui";
import { frame } from "@expo/ui/swift-ui/modifiers";
import { Workspace } from "@/types/workspace";
import { useWorkspaces, useCurrentWorkspace } from "@/lib/hooks";
import { useTheme } from "@/hooks";
import { useAuthStore } from "@/store/auth";
import { useQueryClient } from "@tanstack/react-query";
import { switchWorkspace } from "@/lib/actions/auth";

const WorkspaceItem = ({
  isActive,
  workspace,
  setIsOpened,
}: {
  isActive?: boolean;
  workspace: Workspace;
  setIsOpened: (isOpened: boolean) => void;
}) => {
  const queryClient = useQueryClient();
  const setWorkspace = useAuthStore((state) => state.setWorkspace);

  const handleSwitchWorkspace = async () => {
    if (isActive) {
      setIsOpened(false);
      return;
    }

    queryClient.clear();
    await switchWorkspace(workspace.id);
    setWorkspace(workspace.slug);
    setIsOpened(false);
    await Updates.reloadAsync();
  };

  const { resolvedTheme } = useTheme();
  const mutedTextColor =
    resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300];
  return (
    <HStack spacing={8} onPress={handleSwitchWorkspace}>
      <HStack modifiers={[frame({ width: 38, height: 38 })]}>
        <Avatar
          style={{
            backgroundColor: workspace.avatarUrl ? undefined : workspace.color,
          }}
          name={workspace.name}
          className="size-[38px]"
          rounded="xl"
          src={workspace.avatarUrl}
        />
      </HStack>
      <VStack alignment="leading">
        <Text lineLimit={1} size={15} weight="medium">
          {workspace.name}
        </Text>
        <Text size={14} weight="medium" color={mutedTextColor}>
          {workspace.userRole}
        </Text>
      </VStack>
      <Spacer />
      {isActive && (
        <Image
          systemName="checkmark.circle.fill"
          color={
            resolvedTheme === "light" ? colors.dark.DEFAULT : colors.gray[200]
          }
          size={18}
        />
      )}
    </HStack>
  );
};

export const WorkspaceSwitcher = ({
  isOpened,
  setIsOpened,
}: {
  isOpened: boolean;
  setIsOpened: (isOpened: boolean) => void;
}) => {
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
      <HStack>
        <Text weight="semibold" color={mutedTextColor} size={14}>
          Switch Workspace
        </Text>
      </HStack>
      {workspaces.map((wk) => (
        <WorkspaceItem
          key={wk.id}
          isActive={wk.id === workspace?.id}
          workspace={wk}
          setIsOpened={setIsOpened}
        />
      ))}
    </BottomSheetModal>
  );
};
