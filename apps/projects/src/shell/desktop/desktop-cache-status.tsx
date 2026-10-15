"use client";

import { useSyncExternalStore } from "react";
import { Box, Button, Text } from "ui";
import { useDesktop } from "./desktop-provider";
import { useDesktopOnline } from "./connectivity";
import {
  DESKTOP_CACHE_RETRY_EVENT,
  getDesktopCacheStatus,
  subscribeDesktopCacheStatus,
} from "./cache-status";

/** Recovery belongs in settings; ordinary background downloads stay quiet. */
export function DesktopCacheStatus() {
  const { isDesktop } = useDesktop();
  const status = useSyncExternalStore(
    subscribeDesktopCacheStatus,
    getDesktopCacheStatus,
    () => null,
  );
  const online = useDesktopOnline();
  const error =
    status?.storageError ?? status?.downloadError ?? status?.shellError;
  if (!isDesktop || !error) return null;
  return (
    <Box>
      <Text className="font-medium">Offline storage needs attention</Text>
      <Text className="mt-1" color="muted" role="alert">
        {error}
      </Text>
      <Button
        className="mt-3"
        disabled={!online}
        onClick={() => {
          window.dispatchEvent(new Event(DESKTOP_CACHE_RETRY_EVENT));
        }}
        size="sm"
        variant="outline"
      >
        Retry offline storage
      </Button>
    </Box>
  );
}
