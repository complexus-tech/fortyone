import { useRef, useState } from "react";
import { toast } from "sonner-native";
import { switchWorkspace } from "@/lib/actions/auth";
import { useAuthStore } from "@/store/auth";
import type { Workspace } from "@/types/workspace";

const performWorkspaceSwitch = async (
  workspace: Workspace,
  previousWorkspace: string | null,
  onClose: () => void,
) => {
  try {
    await useAuthStore.getState().setWorkspace(workspace.slug);
    onClose();
    const activeWorkspaceId = await switchWorkspace(workspace.id);
    if (activeWorkspaceId !== workspace.id)
      throw new Error("The server did not confirm the selected workspace.");
  } catch (error) {
    if (
      previousWorkspace &&
      useAuthStore.getState().workspace === workspace.slug
    ) {
      try {
        await useAuthStore.getState().setWorkspace(previousWorkspace);
      } catch {
        // The original error is the actionable failure. Session restoration
        // is retried during the next app bootstrap if this local write fails.
      }
    }
    toast.error("Could not switch workspace", {
      description: error instanceof Error ? error.message : "Please try again.",
    });
  }
};

export const useSwitchWorkspace = (onClose: () => void) => {
  const lock = useRef(false);
  const [pendingId, setPendingId] = useState<string | null>(null);
  const selectWorkspace = (workspace: Workspace): Promise<void> => {
    if (lock.current) return Promise.resolve();
    if (useAuthStore.getState().workspace === workspace.slug) {
      onClose();
      return Promise.resolve();
    }
    lock.current = true;
    setPendingId(workspace.id);
    const previousWorkspace = useAuthStore.getState().workspace;
    return performWorkspaceSwitch(
      workspace,
      previousWorkspace,
      onClose,
    ).finally(() => {
      lock.current = false;
      setPendingId(null);
    });
  };
  return { pendingId, selectWorkspace };
};
