import React from "react";
import { useWindowDimensions } from "react-native";
import { Avatar } from "./avatar";
import { BottomSheetModal } from "./bottom-sheet-modal";
import { colors } from "@/constants";
import {
  Button,
  HStack,
  Image,
  RNHostView,
  ScrollView,
  Spacer,
  Text,
  VStack,
} from "@expo/ui/swift-ui";
import {
  frame,
  font,
  foregroundStyle,
  lineLimit,
  buttonStyle,
  disabled,
  accessibilityLabel,
} from "@expo/ui/swift-ui/modifiers";
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
    <Button
      onPress={handleSwitchWorkspace}
      modifiers={[
        buttonStyle("plain"),
        frame({ minHeight: 44 }),
        disabled(Boolean(pendingId)),
        accessibilityLabel(
          `${workspace.name}, ${isActive ? "current workspace" : workspace.userRole}`,
        ),
      ]}
    >
      <HStack spacing={8}>
        <HStack modifiers={[frame({ width: 38, height: 38 })]}>
          <RNHostView matchContents>
            <Avatar
              style={{
                backgroundColor: workspace.avatarUrl
                  ? undefined
                  : workspace.color,
              }}
              name={workspace.name}
              className="size-[38px]"
              rounded="xl"
              src={workspace.avatarUrl}
            />
          </RNHostView>
        </HStack>
        <VStack alignment="leading">
          <Text
            modifiers={[lineLimit(1), font({ size: 15, weight: "medium" })]}
          >
            {workspace.name}
          </Text>
          <Text
            modifiers={[
              font({ size: 14, weight: "medium" }),
              foregroundStyle(mutedTextColor),
            ]}
          >
            {pendingId === workspace.id ? "Switching…" : workspace.userRole}
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
    </Button>
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
  const { resolvedTheme } = useTheme();
  const { data: workspaces = [] } = useWorkspaces();
  const { workspace } = useCurrentWorkspace();
  const mutedTextColor =
    resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300];
  return (
    <BottomSheetModal
      nativeContent
      isOpen={isOpened}
      onClose={() => setIsOpened(false)}
      spacing={18}
    >
      <HStack>
        <Text
          modifiers={[
            font({ size: 14, weight: "semibold" }),
            foregroundStyle(mutedTextColor),
          ]}
        >
          Switch Workspace
        </Text>
      </HStack>
      <ScrollView
        modifiers={[
          frame({
            height: Math.min(Math.max(workspaces.length, 1) * 64, height * 0.6),
          }),
        ]}
      >
        <VStack spacing={18}>
          {workspaces.map((wk) => (
            <WorkspaceItem
              key={wk.id}
              isActive={wk.id === workspace?.id}
              workspace={wk}
              pendingId={pendingId}
              onSelect={selectWorkspace}
            />
          ))}
        </VStack>
      </ScrollView>
    </BottomSheetModal>
  );
};
