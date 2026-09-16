import type { ContextMenuAction } from "@/components/ui/context-menu.types";
import type { FeedbackItem } from "./types";
import { useRouter } from "expo-router";
import { toast } from "sonner-native";
import { useTerminology } from "@/hooks/use-terminology";
import {
  useDetailAccess,
  useLinkActions,
  confirmAction,
  destructiveColor,
} from "@/modules/entity-details/header";
import { useFeedbackCommands } from "./detail-data";

export type FeedbackDialog = "close" | "link" | "merge" | null;

export function useFeedbackActions(
  item: FeedbackItem,
  teamId: string,
  setDialog: (dialog: FeedbackDialog) => void,
) {
  const router = useRouter();
  const { canEdit, isAdmin, workspace } = useDetailAccess();
  const { getTermDisplay } = useTerminology();
  const commands = useFeedbackCommands(item.id);
  const primary = item.storyLinks?.find((link) => link.isPrimary);
  const inactive = Boolean(item.deletedAt || item.mergedIntoItemId);
  const canTriage = canEdit && !inactive && !primary && !commands.isPending;
  const canPlan = canTriage && item.status !== "closed";
  const canComment = Boolean(workspace) && !inactive;
  const storyTerm = getTermDisplay("storyTerm");
  const linkActions = useLinkActions(
    `/teams/${teamId}/feedback/${item.id}`,
    item.title,
  );
  const openStory = (storyId: string) =>
    router.push({ pathname: "/story/[storyId]", params: { storyId } });
  const prioritize = async () => {
    const result = await commands.execute({ type: "plan", teamId });
    if (result?.storyId) openStory(result.storyId);
  };
  const actions: ContextMenuAction[] = [...linkActions];
  if (!commands.isPending && !inactive)
    actions.push({
      label: item.readAt ? "Mark as unread" : "Mark as read",
      systemImage: "envelope",
      onPress: () => {
        void commands
          .execute({ type: item.readAt ? "unread" : "read" })
          .catch((error: Error) => toast.error(error.message));
      },
    });
  if (canPlan) {
    actions.push({
      label: "Prioritize",
      systemImage: "checkmark.circle",
      onPress: () => {
        void prioritize().catch((error: Error) => toast.error(error.message));
      },
    });
    if (item.status !== "reviewing")
      actions.push({
        label: "Mark as reviewing",
        systemImage: "eye",
        onPress: () => {
          void commands
            .execute({ type: "status", status: "reviewing" })
            .catch((error: Error) => toast.error(error.message));
        },
      });
    actions.push(
      {
        label: `Link existing ${storyTerm}`,
        systemImage: "link",
        onPress: () => setDialog("link"),
      },
      {
        label: "Close feedback",
        systemImage: "xmark.circle",
        onPress: () => setDialog("close"),
      },
    );
  }
  if (primary)
    actions.push({
      label: `Open linked ${storyTerm}`,
      systemImage: "arrow.up.right",
      onPress: () => openStory(primary.storyId),
    });
  if (isAdmin && !primary && !commands.isPending && !item.mergedIntoItemId) {
    if (item.deletedAt)
      actions.push({
        label: "Restore feedback",
        systemImage: "arrow.uturn.backward",
        onPress: () => {
          void commands
            .execute({ type: "restore" })
            .catch((error: Error) => toast.error(error.message));
        },
      });
    else
      actions.push(
        {
          label: "Merge feedback",
          systemImage: "arrow.triangle.merge",
          onPress: () => setDialog("merge"),
        },
        {
          label: "Move to trash",
          systemImage: "trash",
          color: destructiveColor,
          onPress: () =>
            confirmAction(
              "Move to trash",
              "Move this feedback to trash?",
              async () => {
                await commands.execute({ type: "trash" });
                router.back();
              },
              true,
            ),
        },
      );
  }
  return {
    commands,
    actions,
    canPlan,
    canTriage,
    canComment,
    primary,
    openStory,
    prioritize,
  };
}
