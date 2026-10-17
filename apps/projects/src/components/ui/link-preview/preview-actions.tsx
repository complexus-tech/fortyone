"use client";

import { useRef, useState } from "react";
import { LoadingIcon, MoreVerticalIcon, RefreshIcon } from "icons";
import { Button, Menu, Tooltip } from "ui";
import { toast } from "sonner";

export const PreviewActions = ({
  onRefresh,
  refreshing = false,
  refreshedAt,
}: {
  onRefresh: () => Promise<void>;
  refreshing?: boolean;
  refreshedAt?: number;
}) => {
  const pending = useRef(false);
  const [lastRefreshed, setLastRefreshed] = useState<string | null>(null);
  const refresh = () => {
    if (pending.current || refreshing) return;
    pending.current = true;
    void Promise.resolve()
      .then(onRefresh)
      .then(
        () => {
          pending.current = false;
          toast.success("Preview refreshed");
        },
        () => {
          pending.current = false;
          toast.error("Could not refresh preview. Please try again.");
        },
      );
  };

  return (
    <Menu
      onOpenChange={(open) => {
        if (open)
          setLastRefreshed(
            refreshedAt
              ? new Date(refreshedAt).toLocaleString(undefined, {
                  month: "short",
                  day: "numeric",
                  hour: "numeric",
                  minute: "2-digit",
                })
              : null,
          );
      }}
    >
      <Tooltip title={refreshing ? "Refreshing preview…" : "Preview options"}>
        <Menu.Button>
          <Button
            aria-label={refreshing ? "Refreshing preview" : "Preview options"}
            asIcon
            className="size-7 shrink-0 [&_svg]:text-current"
            color="tertiary"
            size="xs"
            type="button"
            variant="naked"
          >
            {refreshing ? (
              <LoadingIcon
                aria-hidden
                className="size-4 motion-safe:animate-spin"
              />
            ) : (
              <MoreVerticalIcon aria-hidden className="size-4" />
            )}
          </Button>
        </Menu.Button>
      </Tooltip>
      <Menu.Items align="end">
        <Menu.Group>
          <Menu.Item disabled={refreshing} onSelect={refresh}>
            <RefreshIcon aria-hidden className="size-4" />
            {refreshing ? "Refreshing preview…" : "Refresh preview"}
          </Menu.Item>
        </Menu.Group>
        {lastRefreshed ? (
          <p className="text-text-muted px-4 pt-1 pb-2">
            Last refreshed {lastRefreshed}
          </p>
        ) : null}
      </Menu.Items>
    </Menu>
  );
};
