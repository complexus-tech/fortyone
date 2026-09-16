import type { ContextMenuAction } from "@/components/ui/context-menu.types";
import { Alert, Share } from "react-native";
import * as Clipboard from "expo-clipboard";
import { toast } from "sonner-native";
import { Back, Row, Text } from "@/components/ui";
import { ContextMenuButton } from "@/components/ui/context-menu-button";
import { useCurrentWorkspace } from "@/lib/hooks";
import { colors } from "@/constants/colors";

export function DetailHeader({
  reference,
  actions = [],
}: {
  reference: string;
  actions?: ContextMenuAction[];
}) {
  return (
    <Row
      style={{ minHeight: 60, paddingVertical: 6, gap: 8 }}
      justify="between"
      align="center"
      asContainer
    >
      <Back />
      <Text
        accessibilityRole="header"
        fontWeight="bold"
        numberOfLines={1}
        style={{ flex: 1, minWidth: 0, fontSize: 17, lineHeight: 20 }}
      >
        {reference}
      </Text>
      {actions.length ? <ContextMenuButton actions={actions} /> : null}
    </Row>
  );
}

export function useDetailAccess() {
  const { workspace } = useCurrentWorkspace();
  return {
    workspace,
    canEdit:
      workspace?.userRole === "admin" || workspace?.userRole === "member",
    isAdmin: workspace?.userRole === "admin",
  };
}

export function useLinkActions(
  path: string,
  title: string,
): ContextMenuAction[] {
  const { workspace } = useCurrentWorkspace();
  if (!workspace) return [];
  const url = `https://${workspace.slug}.fortyone.app${path}`;
  return [
    {
      label: "Copy link",
      systemImage: "link",
      onPress: () => {
        void Clipboard.setStringAsync(url).then(
          () => toast.success("Link copied"),
          () => toast.error("Could not copy link"),
        );
      },
    },
    {
      label: "Share",
      systemImage: "square.and.arrow.up",
      onPress: () => {
        void Share.share({ title, message: `${title}\n${url}`, url }).catch(
          () => toast.error("Could not share this item"),
        );
      },
    },
  ];
}

export function confirmAction(
  label: string,
  message: string,
  onConfirm: () => Promise<unknown>,
  destructive = false,
) {
  Alert.alert(label, message, [
    { text: "Cancel", style: "cancel" },
    {
      text: label,
      style: destructive ? "destructive" : "default",
      onPress: () => {
        void onConfirm().catch((error: unknown) =>
          toast.error(
            error instanceof Error ? error.message : "Please try again.",
          ),
        );
      },
    },
  ]);
}
export const destructiveColor = colors.danger;
