import { useCallback, useEffect, useRef } from "react";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { post } from "@/lib/http";
import { createTypingActivity } from "./typing-activity";

/** Ephemeral presence is best effort; a failed hint never blocks or fails a message. */
export const useTypingActivity = (
  conversationId: string,
  threadRootId?: string,
  enabled = true,
) => {
  const { workspaceSlug } = useWorkspacePath();
  const activity = useRef<ReturnType<typeof createTypingActivity> | undefined>(
    undefined,
  );
  const stop = useCallback(() => {
    activity.current?.stop();
  }, []);
  const changed = useCallback(() => {
    if (!enabled || !conversationId || document.visibilityState !== "visible")
      return;
    activity.current?.changed();
  }, [conversationId, enabled]);
  useEffect(() => {
    const current = createTypingActivity((isTyping) =>
      post(
        `chat/conversations/${conversationId}/typing`,
        { threadRootId, isTyping },
        { workspaceSlug },
        { retry: 0, timeout: 5000 },
      ),
    );
    activity.current = current;
    const hide = () => {
      if (document.visibilityState !== "visible") current.stop();
    };
    document.addEventListener("visibilitychange", hide);
    window.addEventListener("blur", current.stop);
    return () => {
      current.dispose();
      activity.current = undefined;
      document.removeEventListener("visibilitychange", hide);
      window.removeEventListener("blur", current.stop);
    };
  }, [conversationId, threadRootId, workspaceSlug]);
  return { changed, stop };
};
