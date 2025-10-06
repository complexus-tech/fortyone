import React from "react";
import { Avatar, BottomSheetModal } from "@/components/ui";
import { colors } from "@/constants";
import { HStack, Image, Spacer, Text, VStack } from "@expo/ui/swift-ui";
import { frame } from "@expo/ui/swift-ui/modifiers";
import { Workspace } from "@/types/workspace";
import { useWorkspaces, useCurrentWorkspace } from "@/lib/hooks";
import { useColorScheme } from "nativewind";

const WorkspaceItem = ({
  isActive,
  workspace,
}: {
  isActive?: boolean;
  workspace: Workspace;
}) => {
  const { colorScheme } = useColorScheme();
  const mutedTextColor =
    colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300];
  return (
    <HStack spacing={8}>
      <HStack modifiers={[frame({ width: 36, height: 36 })]}>
        <Avatar
          style={{
            backgroundColor: workspace.avatarUrl ? undefined : workspace.color,
          }}
          name={workspace.name}
          className="size-[36px]"
          rounded="xl"
          src={workspace.avatarUrl}
        />
      </HStack>
      <VStack alignment="leading">
        <Text lineLimit={1} size={16}>
          {workspace.name}
        </Text>
        <Text size={16} color={mutedTextColor}>
          {workspace.userRole}
        </Text>
      </VStack>
      <Spacer />
      {isActive && (
        <Image
          systemName="checkmark.circle.fill"
          color={
            colorScheme === "light" ? colors.dark.DEFAULT : colors.gray[200]
          }
          size={20}
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
  const { colorScheme } = useColorScheme();
  const { data: workspaces = [] } = useWorkspaces();
  const { workspace } = useCurrentWorkspace();
  const mutedTextColor =
    colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300];
  return (
    <BottomSheetModal
      isOpen={isOpened}
      onClose={() => setIsOpened(false)}
      spacing={16}
    >
      <HStack>
        <Text weight="semibold" color={mutedTextColor} size={15}>
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
