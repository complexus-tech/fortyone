"use client";

import { useEffect, useRef, useState } from "react";
import { PreviewActions } from "@/components/ui/link-preview/preview-actions";
import { useAttachmentThumbnailCache } from "./thumbnail-cache-context";

export const ThumbnailActions = ({ url }: { url: string }) => {
  const cache = useAttachmentThumbnailCache();
  const request = useRef<AbortController | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [refreshedAt, setRefreshedAt] = useState<number>();
  useEffect(
    () => () => {
      request.current?.abort();
    },
    [],
  );
  if (!cache) return null;
  return (
    <PreviewActions
      refreshing={refreshing}
      refreshedAt={refreshedAt}
      onRefresh={async () => {
        request.current?.abort();
        const controller = new AbortController();
        request.current = controller;
        setRefreshing(true);
        return cache.refresh(url, controller.signal).then(
          (checkedAt) => {
            if (!controller.signal.aborted) {
              setRefreshedAt(checkedAt);
              setRefreshing(false);
            }
          },
          (error: unknown) => {
            if (!controller.signal.aborted) setRefreshing(false);
            throw error;
          },
        );
      }}
    />
  );
};
