"use client";

import { useEffect, useRef } from "react";
import { useSearchParams } from "next/navigation";
import { Badge, Box, Button, Flex, Skeleton, Text } from "ui";
import { toast } from "sonner";
import { useUserRole, useWorkspacePath } from "@/hooks";
import {
  useCreateGoogleDriveConnectSession,
  useDisconnectGoogleDrive,
  useGoogleDriveIntegration,
} from "@/modules/google-drive/hooks";
import { GoogleDriveIcon } from "@/modules/google-drive";
import {
  SectionHeader,
  SettingsBackButton,
} from "@/modules/settings/components";

const getConnectionPresentation = ({
  configured,
  connected,
  requiresReauthorization,
}: {
  configured: boolean;
  connected: boolean;
  requiresReauthorization: boolean;
}) => {
  if (!configured) {
    return {
      badgeColor: "tertiary" as const,
      badgeLabel: "Not connected",
      description: "Google Drive has not been configured for this environment.",
    };
  }
  if (requiresReauthorization) {
    return {
      badgeColor: "warning" as const,
      badgeLabel: "Reconnect",
      description: "Reconnect to restore file access.",
    };
  }
  if (connected) {
    return {
      badgeColor: "success" as const,
      badgeLabel: "Connected",
      description: "Available in this workspace",
    };
  }
  return {
    badgeColor: "tertiary" as const,
    badgeLabel: "Not connected",
    description: "Only files you choose are shared with FortyOne.",
  };
};

export const GoogleDriveSettings = () => {
  const searchParams = useSearchParams();
  const integration = useGoogleDriveIntegration();
  const connect = useCreateGoogleDriveConnectSession();
  const disconnect = useDisconnectGoogleDrive();
  const { userRole } = useUserRole();
  const { withWorkspace } = useWorkspacePath();
  const handledCallbackResult = useRef<string | null>(null);

  useEffect(() => {
    const connected = searchParams.get("google_drive_connected") === "1";
    const connectionError = searchParams.get("google_drive_error");
    if (!connected && !connectionError) return;

    const callbackResult = connected ? "connected" : connectionError;
    if (handledCallbackResult.current === callbackResult) return;
    handledCallbackResult.current = callbackResult;
    if (connected) {
      toast.success("Google Drive connected");
    } else if (connectionError === "access_denied") {
      toast.error("Google Drive connection was cancelled.");
    } else {
      toast.error("Google Drive could not be connected. Please try again.");
    }

    const url = new URL(window.location.href);
    url.searchParams.delete("google_drive_connected");
    url.searchParams.delete("google_drive_error");
    window.history.replaceState({}, "", url.toString());
  }, [searchParams]);

  const startConnection = () => {
    connect.mutate(window.location.href);
  };
  const connectionPresentation = integration.data
    ? getConnectionPresentation(integration.data)
    : null;
  let connectionAction = null;
  if (integration.isError) {
    connectionAction = (
      <Button
        color="tertiary"
        loading={integration.isFetching}
        onClick={() => void integration.refetch()}
        variant="outline"
      >
        Try again
      </Button>
    );
  } else if (
    integration.data?.connected ||
    integration.data?.requiresReauthorization
  ) {
    connectionAction = (
      <Flex gap={2}>
        <Button
          color="tertiary"
          disabled={disconnect.isPending}
          onClick={() => {
            disconnect.mutate();
          }}
          variant="outline"
        >
          Disconnect
        </Button>
        {userRole !== "guest" ? (
          <Button
            color="invert"
            disabled={!integration.data.configured}
            loading={connect.isPending}
            onClick={startConnection}
          >
            Reconnect
          </Button>
        ) : null}
      </Flex>
    );
  } else if (userRole === "guest") {
    connectionAction = null;
  } else {
    connectionAction = (
      <Button
        color="invert"
        disabled={!integration.data?.configured}
        loading={connect.isPending}
        onClick={startConnection}
      >
        Connect Google Drive
      </Button>
    );
  }

  return (
    <Box>
      <Flex align="center" className="mb-6" gap={2}>
        <SettingsBackButton
          href={withWorkspace("/settings/account")}
          label="Back to account settings"
        />
        <Text as="h1" className="text-2xl font-medium">
          Google Drive
        </Text>
      </Flex>

      <Box className="border-border bg-surface overflow-hidden rounded-2xl border">
        <SectionHeader
          action={connectionAction}
          description="Connect the Google account you want to use in this FortyOne workspace."
          title="Personal connection"
        />

        {integration.isPending ? (
          <Box
            aria-label="Loading Google Drive connection"
            className="px-6 py-5"
            role="status"
          >
            <Skeleton className="h-12 w-full" />
          </Box>
        ) : null}
        {integration.isError ? (
          <Box className="px-6 py-6" role="alert">
            <Text fontWeight="semibold">
              Couldn&apos;t load your Google Drive connection
            </Text>
            <Text className="mt-1" color="muted">
              Try again before connecting or changing access.
            </Text>
          </Box>
        ) : null}
        {integration.data ? (
          <Flex align="center" className="px-6 py-5" justify="between">
            <Flex align="center" className="min-w-0" gap={3}>
              <Flex
                align="center"
                className="bg-surface-muted size-10 shrink-0 rounded-xl"
                justify="center"
              >
                <GoogleDriveIcon className="size-6" />
              </Flex>
              <Box className="min-w-0">
                <Text className="truncate" fontWeight="semibold">
                  {integration.data.email ??
                    (integration.data.connected
                      ? "Google account connected"
                      : "No Google account connected")}
                </Text>
                <Text className="truncate" color="muted">
                  {connectionPresentation?.description}
                </Text>
              </Box>
            </Flex>
            <Badge
              color={connectionPresentation?.badgeColor}
              variant={integration.data.connected ? "solid" : "outline"}
            >
              {connectionPresentation?.badgeLabel}
            </Badge>
          </Flex>
        ) : null}
      </Box>
    </Box>
  );
};
