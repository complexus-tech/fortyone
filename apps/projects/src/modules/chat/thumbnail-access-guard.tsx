"use client";

import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useAttachmentThumbnailCache } from "@/components/ui/attachment-preview/thumbnail-cache-context";
import { useWorkspacePath } from "@/hooks/use-workspace-path";

// Pixel caching does not extend a resource's permission grant. Besides the
// thirty-second authorization lease, clear immediately on observed access
// failures and recheck after returning to the application.
export const ThumbnailAccessGuard = () => {
  const cache = useAttachmentThumbnailCache();
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  useEffect(() => {
    if (!cache) return;
    const unsubscribe = queryClient.getQueryCache().subscribe((event) => {
      if (event.type !== "updated" || event.action.type !== "error") return;
      const error = event.query.state.error;
      if (!error || !("status" in error)) return;
      const status = Number(error.status);
      const key = event.query.queryKey;
      if (
        status === 401 ||
        ([403, 404].includes(status) &&
          key[0] === "chat" &&
          key[1] === workspaceSlug.toLowerCase())
      )
        cache.clear();
    });
    const onFocus = () => {
      cache.clear();
    };
    window.addEventListener("focus", onFocus);
    return () => {
      unsubscribe();
      window.removeEventListener("focus", onFocus);
    };
  }, [cache, queryClient, workspaceSlug]);
  return null;
};
