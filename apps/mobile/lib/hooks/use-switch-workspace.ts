import { useRef, useState } from "react";
import { toast } from "sonner-native";
import { switchWorkspace } from "@/lib/actions/auth";
import { useAuthStore } from "@/store/auth";
import type { Workspace } from "@/types/workspace";

export const useSwitchWorkspace = (onClose: () => void) => {
  const lock = useRef(false);
  const [pendingId, setPendingId] = useState<string | null>(null);
  const selectWorkspace = async (workspace: Workspace) => {
    if (lock.current) return;
    if (useAuthStore.getState().workspace === workspace.slug) {
      onClose();
      return;
    }
    lock.current = true;
    setPendingId(workspace.id);
    try {
      await switchWorkspace(workspace.id);
      await useAuthStore.getState().setWorkspace(workspace.slug);
      onClose();
    } catch (error) {
      toast.error("Could not switch workspace", {
        description:
          error instanceof Error ? error.message : "Please try again.",
      });
    } finally {
      lock.current = false;
      setPendingId(null);
    }
  };
  return { pendingId, selectWorkspace };
};
