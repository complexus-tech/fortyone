import React from "react";
import { Avatar, BottomSheetModal } from "@/components/ui";
import { colors } from "@/constants";
import { HStack, Image, Spacer, Text, VStack } from "@expo/ui/swift-ui";
import { frame } from "@expo/ui/swift-ui/modifiers";
import { Workspace } from "@/types/workspace";
import { useWorkspaces, useCurrentWorkspace } from "@/lib/hooks";

const WorkspaceItem = ({
  isActive,
  workspace,
}: {
  isActive?: boolean;
  workspace: Workspace;
}) => {
  return (
    <HStack spacing={8}>
      <HStack modifiers={[frame({ width: 30, height: 30 })]}>
        <Avatar
          style={{
            backgroundColor: workspace.avatarUrl ? undefined : workspace.color,
          }}
          name={workspace.name}
          size="md"
          rounded="lg"
          src={workspace.avatarUrl}
        />
      </HStack>
      <VStack alignment="leading">
        <Text lineLimit={1} size={16}>
          {workspace.name}
        </Text>
        <Text size={16} color={colors.gray.DEFAULT}>
          {workspace.userRole}
        </Text>
      </VStack>
      <Spacer />
      {isActive && (
        <Image systemName="checkmark.circle.fill" color="white" size={20} />
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
  const { data: workspaces = [] } = useWorkspaces();
  const { workspace } = useCurrentWorkspace();
  return (
    <BottomSheetModal isOpen={isOpened} onClose={() => setIsOpened(false)}>
      <HStack>
        <Text weight="semibold" color={colors.gray.DEFAULT} size={15}>
          Switch Workspace
        </Text>
      </HStack>
      {workspaces.map((wk) => (
        <WorkspaceItem
          key={wk.id}
          isActive={wk.id === workspace?.id}
          workspace={wk}
        />
      ))}
    </BottomSheetModal>
  );
};
