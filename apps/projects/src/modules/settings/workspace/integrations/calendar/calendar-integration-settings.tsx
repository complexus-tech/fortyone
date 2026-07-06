"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { useSearchParams } from "next/navigation";
import { format, formatDistanceToNow } from "date-fns";
import { Badge, Box, Button, Dialog, Flex, Menu, Text } from "ui";
import {
  CalendarIcon,
  ClockIcon,
  MoreHorizontalIcon,
  ReloadIcon,
  UnlinkIcon,
} from "icons";
import { useWorkspacePath } from "@/hooks";
import { SectionHeader } from "@/modules/settings/components";
import {
  useCalendarIntegration,
  useCreateCalendarConnectSession,
  useRevokeCalendarConnection,
  useSyncCalendarConnection,
} from "@/lib/hooks/calendar";
import type { CalendarConnection } from "./types";

const GoogleCalendarIcon = ({
  className = "h-5 w-5",
}: {
  className?: string;
}) => (
  <svg
    className={className}
    viewBox="0 0 48 48"
    xmlns="http://www.w3.org/2000/svg"
  >
    <path d="M36 8h5c1.7 0 3 1.3 3 3v5h-8V8Z" fill="#1A73E8" />
    <path d="M4 16h40v21c0 1.7-1.3 3-3 3H7c-1.7 0-3-1.3-3-3V16Z" fill="#fff" />
    <path d="M7 8h29v8H4v-5c0-1.7 1.3-3 3-3Z" fill="#4285F4" />
    <path d="M4 16h40v5H4v-5Z" fill="#E8F0FE" />
    <path
      d="M12.4 32.9c1.3 1.5 3.2 2.3 5.7 2.3 2 0 3.7-.5 4.9-1.6 1.2-1 1.9-2.4 1.9-4.1 0-1.4-.4-2.5-1.2-3.4-.8-.9-1.9-1.4-3.2-1.6l3.9-4v-2.3H12.9v3.1h7l-4 4 .5 2.1h1.7c2 0 3 .7 3 2.1 0 .8-.3 1.4-.9 1.9-.6.5-1.4.7-2.3.7-1.5 0-2.7-.6-3.7-1.7l-1.8 2.5ZM31.2 35h3.6V18.2h-2.7l-5 3.6 1.7 2.6 2.4-1.7V35Z"
      fill="#3C4043"
    />
    <path
      d="M7 40h34c1.7 0 3-1.3 3-3V16h-4v20H7c-1.7 0-3-1.3-3-3v4c0 1.7 1.3 3 3 3Z"
      fill="#34A853"
      opacity=".16"
    />
  </svg>
);

const getSyncBadgeColor = (status?: string) => {
  if (status === "failed") return "warning";
  if (status === "synced" || status === "connected") return "success";
  return "tertiary";
};

const getDetailsBadgeColor = (connection?: CalendarConnection) =>
  connection?.canReadEventDetails ? "success" : "warning";

const formatSyncedAt = (value?: string | null) => {
  if (!value) return "Not synced yet";
  return formatDistanceToNow(new Date(value), { addSuffix: true });
};

const formatExactDate = (value?: string | null) => {
  if (!value) return "Not available";
  return format(new Date(value), "MMM d, yyyy 'at' h:mm a");
};

export const CalendarIntegrationSettings = () => {
  const searchParams = useSearchParams();
  const { data: integration } = useCalendarIntegration();
  const { withWorkspace } = useWorkspacePath();
  const createConnectSession = useCreateCalendarConnectSession();
  const syncConnection = useSyncCalendarConnection();
  const revokeConnection = useRevokeCalendarConnection();
  const [disconnectConnection, setDisconnectConnection] =
    useState<CalendarConnection | null>(null);

  const connection = useMemo(
    () => integration?.connections.find((item) => item.provider === "google"),
    [integration?.connections],
  );

  useEffect(() => {
    if (searchParams.get("connected") !== "1") {
      return;
    }
    const url = new URL(window.location.href);
    url.searchParams.delete("connected");
    window.history.replaceState({}, "", url.toString());
  }, [searchParams]);

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Calendar
      </Text>

      <Box className="border-border bg-surface rounded-2xl border">
        <SectionHeader
          action={
            <Button
              color="invert"
              loading={createConnectSession.isPending}
              onClick={() => {
                createConnectSession.mutate();
              }}
            >
              {connection
                ? "Reconnect Google Calendar"
                : "Connect Google Calendar"}
            </Button>
          }
          description="Connect Google Calendar so FortyOne can understand real availability before recommending schedules and deadlines."
          title="Connected calendar"
        />

        {!connection ? (
          <Box className="px-6 py-8">
            <Text className="font-medium">No calendar connected</Text>
            <Text className="mt-1" color="muted">
              Connect Google Calendar to sync busy/free availability for future
              schedule planning.
            </Text>
          </Box>
        ) : (
          <Flex align="center" className="px-6 py-4" justify="between">
            <Flex align="center" gap={3}>
              <Flex
                align="center"
                className="bg-surface-muted size-9 shrink-0 rounded-lg"
                justify="center"
              >
                <GoogleCalendarIcon className="h-5 w-5" />
              </Flex>
              <Box>
                <Text className="font-medium">{connection.connectedEmail}</Text>
                <Text color="muted">
                  Last synced {formatSyncedAt(connection.lastSyncedAt)}
                </Text>
              </Box>
            </Flex>
            <Flex align="center" gap={2}>
              <Badge
                className="uppercase"
                color={getSyncBadgeColor(connection.syncStatus)}
              >
                {connection.syncStatus}
              </Badge>
              <Badge
                className="uppercase"
                color={getDetailsBadgeColor(connection)}
              >
                {connection.canReadEventDetails ? "Event details" : "Busy only"}
              </Badge>
              <Menu>
                <Menu.Button>
                  <Button
                    className="px-2"
                    color="tertiary"
                    leftIcon={<MoreHorizontalIcon />}
                  />
                </Menu.Button>
                <Menu.Items align="end">
                  <Menu.Group>
                    <Menu.Item
                      onSelect={() => {
                        syncConnection.mutate(connection.id);
                      }}
                    >
                      <ReloadIcon />
                      Sync availability
                    </Menu.Item>
                    <Menu.Item
                      onSelect={() => {
                        createConnectSession.mutate();
                      }}
                    >
                      <GoogleCalendarIcon className="h-4 w-4" />
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
        )}
      </Box>

      <Box className="border-border bg-surface mt-6 rounded-2xl border">
        <SectionHeader
          description="FortyOne stores synced time ranges and visible event titles from your primary calendar. Private events stay hidden, and guests, notes, locations, and descriptions are not stored."
          title="Calendar data"
        />
        <Box className="grid grid-cols-1 gap-3 px-6 py-5 md:grid-cols-2">
          <Flex
            align="center"
            className="border-border bg-surface-muted rounded-xl border px-4 py-3"
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
            className="border-border bg-surface-muted rounded-xl border px-4 py-3"
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

      <Box className="mt-6">
        <Link href={withWorkspace("/settings/workspace/integrations")}>
          <Text color="muted">Back to integrations</Text>
        </Link>
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
              FortyOne will stop syncing availability from this Google Calendar
              connection. Existing busy windows will stop being refreshed.
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
