"use client";

import { useEffect, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import { format, formatDistanceToNow } from "date-fns";
import { toast } from "sonner";
import { Badge, Box, Button, Dialog, Flex, Menu, Skeleton, Text } from "ui";
import {
  CalendarIcon,
  CalendarPlusIcon,
  ClockIcon,
  GoogleCalendarIcon,
  MoreHorizontalIcon,
  ReloadIcon,
  UnlinkIcon,
} from "icons";
import { MicrosoftIcon } from "@/components/ui";
import { useWorkspacePath } from "@/hooks";
import {
  SectionHeader,
  SettingsBackButton,
} from "@/modules/settings/components";
import {
  useCalendarIntegration,
  useCreateCalendarConnectSession,
  useRevokeCalendarConnection,
  useSetPrimaryCalendarConnection,
  useSyncCalendarConnection,
} from "@/lib/hooks/calendar";
import type { CalendarConnection, CalendarProvider } from "./types";

const providers: {
  id: CalendarProvider;
  name: string;
  description: string;
}[] = [
  {
    id: "google",
    name: "Google Calendar",
    description: "Sync meetings, availability, and FortyOne scheduled work.",
  },
  {
    id: "microsoft",
    name: "Outlook Calendar",
    description: "Sync meetings, availability, and FortyOne scheduled work.",
  },
];

const ProviderIcon = ({ provider }: { provider: CalendarProvider }) =>
  provider === "google" ? (
    <GoogleCalendarIcon aria-hidden="true" className="h-6 w-6" />
  ) : (
    <MicrosoftIcon />
  );

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

const getConnectionUsage = (connection: CalendarConnection) => {
  if (connection.isPrimary) return "FortyOne scheduled work";
  if (connection.canWriteEvents) return "Availability only";
  return "Read only";
};

export const CalendarIntegrationSettings = () => {
  const searchParams = useSearchParams();
  const integrationQuery = useCalendarIntegration();
  const connections = integrationQuery.data?.connections ?? [];
  const { withWorkspace } = useWorkspacePath();
  const createConnectSession = useCreateCalendarConnectSession();
  const syncConnection = useSyncCalendarConnection();
  const setPrimaryConnection = useSetPrimaryCalendarConnection();
  const revokeConnection = useRevokeCalendarConnection();
  const [disconnectConnection, setDisconnectConnection] =
    useState<CalendarConnection | null>(null);
  const handledCallbackResult = useRef<string | null>(null);

  useEffect(() => {
    const connected = searchParams.get("connected") === "1";
    const calendarError = searchParams.get("calendar_error");
    const provider = searchParams.get("calendar_provider");
    if (!connected && !calendarError) return;

    const callbackResult = `${provider}:${connected ? "connected" : calendarError}`;
    if (handledCallbackResult.current === callbackResult) return;
    handledCallbackResult.current = callbackResult;
    const providerName =
      provider === "microsoft" ? "Outlook Calendar" : "Google Calendar";
    if (!connected && calendarError === "access_denied") {
      toast.error(`${providerName} connection was cancelled.`);
    } else if (!connected) {
      toast.error(`${providerName} could not be connected. Please try again.`);
    }
    const url = new URL(window.location.href);
    url.searchParams.delete("connected");
    url.searchParams.delete("calendar_error");
    url.searchParams.delete("calendar_provider");
    window.history.replaceState({}, "", url.toString());
  }, [searchParams]);

  const latestSync = connections
    .map((connection) => connection.lastSyncedAt)
    .filter((value): value is string => Boolean(value))
    .toSorted((left, right) => Date.parse(right) - Date.parse(left))[0];

  return (
    <Box>
      <Flex align="center" className="mb-6" gap={2}>
        <SettingsBackButton
          href={withWorkspace("/settings/account")}
          label="Back to account settings"
        />
        <Text as="h1" className="text-2xl font-medium">
          Calendar
        </Text>
      </Flex>

      <Box className="border-border bg-surface rounded-2xl border">
        <SectionHeader
          action={
            integrationQuery.isError ? (
              <Button
                color="tertiary"
                loading={integrationQuery.isFetching}
                onClick={() => void integrationQuery.refetch()}
                variant="outline"
              >
                Try again
              </Button>
            ) : null
          }
          description="Connect calendars for availability, then choose where FortyOne writes scheduled work."
          title="Calendar connections"
        />

        {integrationQuery.isPending ? (
          <Box
            aria-label="Loading calendar connections"
            className="px-6 py-8"
            role="status"
          >
            <Skeleton className="h-12 w-full" />
            <Skeleton className="mt-3 h-12 w-full" />
          </Box>
        ) : null}
        {integrationQuery.isError ? (
          <Box className="px-6 py-8" role="alert">
            <Text className="font-medium">
              Couldn&apos;t load your calendar connections
            </Text>
            <Text className="mt-1" color="muted">
              Try again before connecting or changing calendar access.
            </Text>
          </Box>
        ) : null}
        {!integrationQuery.isPending && !integrationQuery.isError
          ? providers.map((provider, index) => {
              const connection = connections.find(
                (item) => item.provider === provider.id,
              );
              const status = getConnectionStatus(connection?.syncStatus);
              const isConnecting =
                createConnectSession.isPending &&
                createConnectSession.variables === provider.id;
              const isMakingPrimary =
                setPrimaryConnection.isPending &&
                setPrimaryConnection.variables === connection?.id;
              return (
                <Box
                  className={index > 0 ? "border-border border-t" : undefined}
                  key={provider.id}
                >
                  {connection?.requiresReauthorization ? (
                    <Flex
                      align="center"
                      className="border-border bg-surface-prominent/35 gap-4 border-b px-6 py-4"
                      justify="between"
                      role="status"
                    >
                      <Flex align="center" className="min-w-0" gap={3}>
                        <CalendarPlusIcon
                          aria-hidden="true"
                          className="h-5 w-auto"
                        />
                        <Box className="min-w-0">
                          <Text className="font-medium">
                            Reconnect to schedule work on {provider.name}
                          </Text>
                          <Text className="mt-0.5" color="muted">
                            Reconnect once to restore event updates and cleanup.
                          </Text>
                        </Box>
                      </Flex>
                      <Button
                        color="tertiary"
                        loading={isConnecting}
                        onClick={() => {
                          createConnectSession.mutate(provider.id);
                        }}
                        variant="outline"
                      >
                        Reconnect
                      </Button>
                    </Flex>
                  ) : null}
                  <Flex
                    align="center"
                    className="gap-4 px-6 py-4"
                    justify="between"
                  >
                    <Flex align="center" className="min-w-0" gap={3}>
                      <Flex
                        align="center"
                        className="bg-surface-muted size-10 shrink-0 rounded-xl"
                        justify="center"
                      >
                        <ProviderIcon provider={provider.id} />
                      </Flex>
                      <Box className="min-w-0">
                        <Text className="truncate font-medium">
                          {connection?.connectedEmail ?? provider.name}
                        </Text>
                        <Text className="truncate" color="muted">
                          {connection
                            ? `${getConnectionUsage(connection)} · ${formatSyncedAt(connection.lastSyncedAt)}`
                            : provider.description}
                        </Text>
                      </Box>
                    </Flex>
                    <Flex align="center" className="shrink-0" gap={2}>
                      {connection?.isPrimary ? (
                        <Badge color="tertiary" variant="solid">
                          Primary
                        </Badge>
                      ) : null}
                      {status ? (
                        <Badge color={status.color} variant={status.variant}>
                          {status.label}
                        </Badge>
                      ) : null}
                      {!connection ? (
                        <Button
                          color="invert"
                          loading={isConnecting}
                          onClick={() => {
                            createConnectSession.mutate(provider.id);
                          }}
                        >
                          Connect
                        </Button>
                      ) : (
                        <>
                          {!connection.isPrimary &&
                          connection.canWriteEvents ? (
                            <Button
                              color="tertiary"
                              loading={isMakingPrimary}
                              onClick={() => {
                                setPrimaryConnection.mutate(connection.id);
                              }}
                              variant="outline"
                            >
                              Make primary
                            </Button>
                          ) : null}
                          <Menu>
                            <Menu.Button>
                              <Button
                                aria-label={`${provider.name} connection actions`}
                                className="px-2"
                                color="tertiary"
                                leftIcon={<MoreHorizontalIcon />}
                                variant="naked"
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
                                  Sync calendar
                                </Menu.Item>
                                <Menu.Item
                                  onSelect={() => {
                                    createConnectSession.mutate(provider.id);
                                  }}
                                >
                                  <ProviderIcon provider={provider.id} />
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
                        </>
                      )}
                    </Flex>
                  </Flex>
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
              );
            })
          : null}
      </Box>

      <Box className="border-border bg-surface mt-6 rounded-2xl border">
        <SectionHeader
          description="FortyOne keeps an owner-only cache of connected calendars. Private events remain Busy; teammates and managers receive availability only."
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
              <Text color="muted">{formatExactDate(latestSync)}</Text>
            </Box>
          </Flex>
        </Box>
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
              FortyOne will stop syncing this calendar, remove its imported
              events, and clean up FortyOne work events written to it.
            </Text>
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-3 border-0 pt-2">
            <Button
              color="tertiary"
              onClick={() => {
                setDisconnectConnection(null);
              }}
            >
              Cancel
            </Button>
            <Button
              loading={revokeConnection.isPending}
              onClick={() => {
                if (!disconnectConnection) return;
                revokeConnection.mutate(disconnectConnection.id, {
                  onSuccess: (res) => {
                    if (!res.error) setDisconnectConnection(null);
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
