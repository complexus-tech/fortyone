import type { Workspace } from "@/types/workspace";
import React from "react";
import { useWindowDimensions } from "react-native";
import { Avatar } from "./Avatar";
import { BottomSheetModal } from "./bottom-sheet-modal";
import { themeColors } from "@/constants/colors";
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
  padding,
  contentShape,
  shapes,
} from "@expo/ui/swift-ui/modifiers";
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
  const mutedTextColor = themeColors[resolvedTheme].textMuted;
  return (
    <Button
      onPress={handleSwitchWorkspace}
      modifiers={[
        buttonStyle("plain"),
        frame({ minHeight: 68 }),
        disabled(Boolean(pendingId)),
        accessibilityLabel(
          `${workspace.name}, ${isActive ? "current workspace" : workspace.userRole}`,
        ),
      ]}
    >
      <HStack
        spacing={12}
        modifiers={[padding({ vertical: 8 }), contentShape(shapes.rectangle())]}
      >
        <HStack modifiers={[frame({ width: 40, height: 40 })]}>
          <RNHostView matchContents>
            <Avatar
              style={{
                backgroundColor: workspace.avatarUrl
                  ? undefined
                  : workspace.color,
              }}
              name={workspace.name}
              className="size-[40px]"
              rounded="xl"
              src={workspace.avatarUrl}
            />
          </RNHostView>
        </HStack>
        <VStack alignment="leading" spacing={3}>
          <Text
            modifiers={[
              lineLimit(1),
              font({ textStyle: "callout", weight: "medium" }),
            ]}
          >
            {workspace.name}
          </Text>
          <Text
            modifiers={[
              lineLimit(1),
              font({ textStyle: "subheadline", weight: "regular" }),
              foregroundStyle(mutedTextColor),
            ]}
          >
            {pendingId === workspace.id
              ? "Switching…"
              : `${workspace.slug} · ${workspace.userRole}`}
          </Text>
        </VStack>
        <Spacer />
        {isActive && (
          <Image
            systemName="checkmark"
            color={themeColors[resolvedTheme].foreground}
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
  const mutedTextColor = themeColors[resolvedTheme].textMuted;
  return (
    <BottomSheetModal
      nativeContent
      isOpen={isOpened}
      onClose={() => setIsOpened(false)}
      spacing={16}
      padding={{ leading: 20, trailing: 20, top: 32, bottom: 16 }}
    >
      <VStack alignment="leading" spacing={6}>
        <Text
          modifiers={[
            lineLimit(1),
            font({ textStyle: "title3", weight: "semibold" }),
          ]}
        >
          Choose a workspace
        </Text>
        <Text
          modifiers={[
            font({ textStyle: "callout" }),
            foregroundStyle(mutedTextColor),
          ]}
        >
          Select the workspace you want to use.
        </Text>
      </VStack>
      <ScrollView
        modifiers={[
          frame({
            height: Math.min(Math.max(workspaces.length, 1) * 72, height * 0.6),
          }),
        ]}
      >
        <VStack spacing={4}>
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
