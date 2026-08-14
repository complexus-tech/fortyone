"use client";

import { useEffect, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import { format, formatDistanceToNow } from "date-fns";
import { toast } from "sonner";
import { Badge, Box, Button, Dialog, Flex, Menu, Skeleton, Text } from "ui";
import {
  CalendarIcon,
  ClockIcon,
  GoogleCalendarIcon,
  MoreHorizontalIcon,
  ReloadIcon,
  UnlinkIcon,
} from "icons";
import { useWorkspacePath } from "@/hooks";
import {
  SectionHeader,
  SettingsBackButton,
} from "@/modules/settings/components";
import {
  useCalendarIntegration,
  useCreateCalendarConnectSession,
  useRevokeCalendarConnection,
  useSyncCalendarConnection,
} from "@/lib/hooks/calendar";
import type { CalendarConnection } from "./types";

const getConnectionStatus = (status?: string) => {
  if (status === "failed") {
    return {
      color: "warning" as const,
      label: "Needs attention",
      variant: "outline" as const,
    };
  }
  if (status === "syncing") {
    return {
      color: "tertiary" as const,
      label: "Syncing",
      variant: "solid" as const,
    };
  }
  return null;
};

const formatSyncedAt = (value?: string | null) => {
  if (!value) return "Waiting for first sync";
  return `Synced ${formatDistanceToNow(new Date(value), { addSuffix: true })}`;
};

const formatExactDate = (value?: string | null) => {
  if (!value) return "Not available";
  return format(new Date(value), "MMM d, yyyy 'at' h:mm a");
};

export const CalendarIntegrationSettings = () => {
  const searchParams = useSearchParams();
  const integrationQuery = useCalendarIntegration();
  const integration = integrationQuery.data;
  const { withWorkspace } = useWorkspacePath();
  const createConnectSession = useCreateCalendarConnectSession();
  const syncConnection = useSyncCalendarConnection();
  const revokeConnection = useRevokeCalendarConnection();
  const [disconnectConnection, setDisconnectConnection] =
    useState<CalendarConnection | null>(null);
  const handledCallbackResult = useRef<string | null>(null);

  const connection = integration?.connections.find(
    (item) => item.provider === "google",
  );
  const connectionStatus = getConnectionStatus(connection?.syncStatus);

  useEffect(() => {
    const connected = searchParams.get("connected") === "1";
    const calendarError = searchParams.get("calendar_error");
    if (!connected && !calendarError) {
      return;
    }
    const callbackResult = connected ? "connected" : calendarError;
    if (handledCallbackResult.current === callbackResult) {
      return;
    }
    handledCallbackResult.current = callbackResult;
    if (connected) {
      toast.success("Google Calendar connected");
    } else if (calendarError === "access_denied") {
      toast.error("Google Calendar connection was cancelled.");
    } else {
      toast.error("Google Calendar could not be connected. Please try again.");
    }
    const url = new URL(window.location.href);
    url.searchParams.delete("connected");
    url.searchParams.delete("calendar_error");
    window.history.replaceState({}, "", url.toString());
  }, [searchParams]);

  return (
    <Box>
      <Flex align="center" className="mb-6" gap={2}>
        <SettingsBackButton
          href={withWorkspace("/settings/account")}
          label="Back to account settings"
        />
        <Text as="h1" className="text-2xl font-medium">
          Google Calendar
        </Text>
      </Flex>

      <Box className="border-border bg-surface rounded-2xl border">
        <SectionHeader
          action={
            <Flex className="shrink-0">
              {integrationQuery.isPending ? (
                <Skeleton className="h-10 w-48" />
              ) : null}
              {integrationQuery.isError ? (
                <Button
                  className="shrink-0 whitespace-nowrap"
                  color="tertiary"
                  loading={integrationQuery.isFetching}
                  onClick={() => {
                    void integrationQuery.refetch();
                  }}
                  variant="outline"
                >
                  Try again
                </Button>
              ) : null}
              {!integrationQuery.isPending && !integrationQuery.isError ? (
                <Button
                  className="shrink-0 whitespace-nowrap"
                  color="invert"
                  leftIcon={
                    <GoogleCalendarIcon
                      aria-hidden="true"
                      className="h-4.5 w-4.5 shrink-0"
                    />
                  }
                  loading={createConnectSession.isPending}
                  onClick={() => {
                    createConnectSession.mutate();
                  }}
                >
                  {connection ? "Reconnect" : "Connect"}
                </Button>
              ) : null}
            </Flex>
          }
          description="Connect your primary Google Calendar to see meetings alongside FortyOne work and help future plans respect that calendar's availability."
          title="Calendar"
        />

        {integrationQuery.isPending ? (
          <Box
            aria-label="Loading calendar connection"
            className="px-6 py-8"
            role="status"
          >
            <Skeleton className="h-5 w-48" />
            <Skeleton className="mt-3 h-4 w-96 max-w-full" />
          </Box>
        ) : null}
        {integrationQuery.isError ? (
          <Box className="px-6 py-8" role="alert">
            <Text className="font-medium">
              Couldn&apos;t load your calendar connection
            </Text>
            <Text className="mt-1" color="muted">
              Try again before connecting or changing calendar access.
            </Text>
          </Box>
        ) : null}
        {!integrationQuery.isPending &&
        !integrationQuery.isError &&
        !connection ? (
          <Box className="px-6 py-8">
            <Text className="font-medium">No calendar connected</Text>
            <Text className="mt-1" color="muted">
              Connect your primary Google Calendar to bring its meetings into
              Calendar and keep scheduled work clear of those commitments.
            </Text>
          </Box>
        ) : null}
        {!integrationQuery.isPending &&
        !integrationQuery.isError &&
        connection ? (
          <Flex align="center" className="gap-4 px-6 py-4" justify="between">
            <Flex align="center" className="min-w-0" gap={3}>
              <Flex
                align="center"
                className="bg-surface-muted size-10 shrink-0 rounded-xl"
                justify="center"
              >
                <GoogleCalendarIcon aria-hidden="true" className="h-6 w-6" />
              </Flex>
              <Box className="min-w-0">
                <Text className="truncate font-medium">
                  {connection.connectedEmail}
                </Text>
                <Text className="truncate" color="muted">
                  Primary calendar ·{" "}
                  {connection.canReadEventDetails
                    ? "Event details enabled"
                    : "Availability only"}
                  {" · "}
                  {formatSyncedAt(connection.lastSyncedAt)}
                </Text>
              </Box>
            </Flex>
            <Flex align="center" className="shrink-0" gap={2}>
              {connectionStatus ? (
                <Badge
                  color={connectionStatus.color}
                  variant={connectionStatus.variant}
                >
                  {connectionStatus.label}
                </Badge>
              ) : null}
              <Menu>
                <Menu.Button>
                  <Button
                    aria-label="Calendar connection actions"
                    className="px-2"
                    color="tertiary"
                    leftIcon={<MoreHorizontalIcon />}
                  />
                </Menu.Button>
                <Menu.Items align="end">
                  <Menu.Group>
                    <Menu.Item
                      onSelect={() => {
                        syncConnection.mutate({
                          connectionId: connection.id,
                        });
                      }}
                    >
                      <ReloadIcon />
                      {connection.canReadEventDetails
                        ? "Sync calendar"
                        : "Sync availability"}
                    </Menu.Item>
                    <Menu.Item
                      onSelect={() => {
                        createConnectSession.mutate();
                      }}
                    >
                      <GoogleCalendarIcon
                        aria-hidden="true"
                        className="h-4 w-4"
                      />
                      Update connection
                    </Menu.Item>
                    <Menu.Item
                      className="text-danger"
                      onSelect={() => {
                        setDisconnectConnection(connection);
                      }}
                    >
                      <UnlinkIcon className="text-danger" />
                      Disconnect calendar
                    </Menu.Item>
                  </Menu.Group>
                </Menu.Items>
              </Menu>
            </Flex>
          </Flex>
        ) : null}
      </Box>

      <Box className="border-border bg-surface mt-6 rounded-2xl border">
        <SectionHeader
          description="FortyOne keeps an owner-only cache of your primary calendar so Calendar can show titles, locations, meeting links, descriptions, and attendees. Private events remain Busy; teammates and managers receive availability only."
          title="Calendar data"
        />
        <Box className="grid grid-cols-1 gap-3 px-6 py-5 md:grid-cols-2">
          <Flex
            align="center"
            className="border-border bg-surface-prominent/45 rounded-xl border px-4 py-3"
            gap={3}
          >
            <CalendarIcon className="h-5 w-auto" />
            <Box>
              <Text className="font-medium">Sync window</Text>
              <Text color="muted">7 days back, 90 days ahead</Text>
            </Box>
          </Flex>
          <Flex
            align="center"
            className="border-border bg-surface-prominent/45 rounded-xl border px-4 py-3"
            gap={3}
          >
            <ClockIcon className="h-5 w-auto" />
            <Box>
              <Text className="font-medium">Latest sync</Text>
              <Text color="muted">
                {formatExactDate(connection?.lastSyncedAt)}
              </Text>
            </Box>
          </Flex>
        </Box>
        {connection?.syncError ? (
          <Box className="border-border border-t px-6 py-4">
            <Text className="font-medium" color="danger">
              Sync failed
            </Text>
            <Text className="mt-1" color="muted">
              {connection.syncError}
            </Text>
          </Box>
        ) : null}
      </Box>

      <Dialog
        onOpenChange={(open) => {
          if (!open) setDisconnectConnection(null);
        }}
        open={Boolean(disconnectConnection)}
      >
        <Dialog.Content>
          <Dialog.Header>
            <Dialog.Title className="px-6 pt-0.5 text-lg">
              Disconnect calendar
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Text color="muted">
              FortyOne will stop syncing this primary Google Calendar and remove
              its events from your FortyOne calendar.
            </Text>
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-3 border-0 pt-2">
            <Button
              className="px-4"
              color="tertiary"
              onClick={() => {
                setDisconnectConnection(null);
              }}
            >
              Cancel
            </Button>
            <Button
              className="px-4"
              loading={revokeConnection.isPending}
              onClick={() => {
                if (!disconnectConnection) return;
                revokeConnection.mutate(disconnectConnection.id, {
                  onSuccess: (res) => {
                    if (!res.error) {
                      setDisconnectConnection(null);
                    }
                  },
                });
              }}
            >
              Disconnect
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </Box>
  );
};
