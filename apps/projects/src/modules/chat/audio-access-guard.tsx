"use client";

import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { getMessage } from "./api";
import { useAudioPlaybackController } from "./audio-playback-context";

// A virtualized row may have unmounted while its audio is still playing.
// Revalidate that source when chat receives a change hint, including deletion.
export const AudioAccessGuard = () => {
  const client = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const playback = useAudioPlaybackController();
  useEffect(() => {
    let timer: ReturnType<typeof setTimeout> | undefined;
    let request: AbortController | undefined;
    let needsCheck = false;
    const schedule = () => {
      if (timer || request || !needsCheck) return;
      timer = setTimeout(() => {
        timer = undefined;
        needsCheck = false;
        const track = playback.getSnapshot().track;
        if (!track) return;
        const controller = new AbortController();
        request = controller;
        const isCurrentTrack = () => {
          const current = playback.getSnapshot().track;
          return (
            !controller.signal.aborted &&
            current?.id === track.id &&
            current.messageId === track.messageId
          );
        };
        void getMessage({ workspaceSlug }, track.messageId, controller.signal)
          .then((message) => {
            if (!isCurrentTrack()) return;
            if (
              message.deletedAt ||
              !message.attachments.some(({ id }) => id === track.id)
            )
              playback.invalidateMessage(track.messageId);
          })
          .catch((error: unknown) => {
            if (!isCurrentTrack()) return;
            const status =
              error && typeof error === "object" && "status" in error
                ? Number(error.status)
                : 0;
            if ([401, 403, 404].includes(status))
              playback.invalidateMessage(track.messageId);
          })
          .finally(() => {
            if (controller.signal.aborted) return;
            request = undefined;
            schedule();
          });
      }, 200);
    };
    const unsubscribe = client.getQueryCache().subscribe((event) => {
      if (event.type !== "updated" || event.action.type !== "invalidate")
        return;
      const key = event.query.queryKey;
      if (
        key[0] !== "chat" ||
        key[1] !== workspaceSlug.toLowerCase() ||
        !playback.getSnapshot().track
      )
        return;
      needsCheck = true;
      schedule();
    });
    return () => {
      unsubscribe();
      if (timer) clearTimeout(timer);
      request?.abort();
    };
  }, [client, playback, workspaceSlug]);
  return null;
};
