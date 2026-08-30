import { format } from "date-fns";
import { CalendarIcon, GoogleCalendarIcon } from "icons";
import { Box, Button, Flex, Text } from "ui";
import { MicrosoftIcon } from "@/components/ui/microsoft-icon";

type CalendarNoticeConnection = {
  id: string;
  lastSyncedAt?: string | null;
  provider: string;
  requiresReauthorization?: boolean;
  syncError?: string | null;
  syncStatus?: string | null;
};

const getCalendarProviderName = (connection?: CalendarNoticeConnection) =>
  connection?.provider === "microsoft" ? "Outlook Calendar" : "Google Calendar";

const CalendarProviderIcon = ({
  connection,
}: {
  connection?: CalendarNoticeConnection;
}) =>
  connection?.provider === "microsoft" ? (
    <MicrosoftIcon />
  ) : (
    <GoogleCalendarIcon aria-hidden="true" className="h-6 w-6" />
  );

export const CalendarNotices = ({
  canReadEventDetails,
  conflictCount,
  connectHref,
  connection,
  hasIntegrationError,
  isIntegrationPending,
  isReconnectPending,
  isSyncing,
  onReconnect,
  onSync,
}: {
  canReadEventDetails: boolean;
  conflictCount: number;
  connectHref: string;
  connection?: CalendarNoticeConnection;
  hasIntegrationError: boolean;
  isIntegrationPending: boolean;
  isReconnectPending: boolean;
  isSyncing: boolean;
  onReconnect: () => void;
  onSync: (connectionId: string) => void;
}) => (
  <>
    {!isIntegrationPending && !hasIntegrationError && !connection ? (
      <Box className="border-border border-b px-5 py-3">
        <Flex align="center" gap={3} justify="between">
          <Flex align="center" className="min-w-0" gap={3}>
            <Box className="flex h-10 w-10 shrink-0 items-center justify-center">
              <CalendarIcon aria-hidden="true" className="h-6 w-6" />
            </Box>
            <Box className="min-w-0">
              <Text fontSize="md" fontWeight="medium">
                Calendar is not connected
              </Text>
              <Text className="line-clamp-1" color="muted" fontSize="md">
                FortyOne can still schedule work blocks, but availability will
                be incomplete until you connect your primary calendar.
              </Text>
            </Box>
          </Flex>
          <Button
            className="text-base"
            color="tertiary"
            href={connectHref}
            variant="outline"
          >
            Connect
          </Button>
        </Flex>
      </Box>
    ) : null}

    {connection?.requiresReauthorization ? (
      <Box className="border-border bg-surface-prominent/30 border-b px-5 py-3">
        <Flex align="center" gap={3} justify="between">
          <Flex align="center" className="min-w-0" gap={3}>
            <Box className="flex h-10 w-10 shrink-0 items-center justify-center">
              <CalendarProviderIcon connection={connection} />
            </Box>
            <Box className="min-w-0">
              <Text fontSize="md" fontWeight="medium">
                Reconnect to update {getCalendarProviderName(connection)} work
                blocks
              </Text>
              <Text className="line-clamp-1" color="muted" fontSize="md">
                Reconnect once to let FortyOne add and update scheduled work in
                your primary calendar.
              </Text>
            </Box>
          </Flex>
          <Button
            className="text-base"
            color="tertiary"
            loading={isReconnectPending}
            onClick={onReconnect}
            variant="outline"
          >
            Reconnect
          </Button>
        </Flex>
      </Box>
    ) : null}

    {connection &&
    !connection.requiresReauthorization &&
    !canReadEventDetails ? (
      <Box className="border-border border-b px-5 py-3">
        <Flex align="center" gap={3} justify="between">
          <Flex align="center" className="min-w-0" gap={3}>
            <Box className="flex h-10 w-10 shrink-0 items-center justify-center">
              <CalendarProviderIcon connection={connection} />
            </Box>
            <Box className="min-w-0">
              <Text fontSize="md" fontWeight="medium">
                Event details are not enabled
              </Text>
              <Text className="line-clamp-1" color="muted" fontSize="md">
                Reconnect your primary {getCalendarProviderName(connection)} to
                show event titles instead of availability-only busy blocks.
              </Text>
            </Box>
          </Flex>
          <Button
            className="text-base"
            color="tertiary"
            loading={isReconnectPending}
            onClick={onReconnect}
            variant="outline"
          >
            Reconnect
          </Button>
        </Flex>
      </Box>
    ) : null}

    {connection?.syncStatus === "failed" ? (
      <Box className="border-warning/30 border-b px-5 py-3">
        <Flex align="center" gap={3} justify="between">
          <Box className="min-w-0">
            <Text fontSize="md" fontWeight="medium">
              Calendar sync failed
            </Text>
            <Text color="muted" fontSize="md">
              {connection.syncError?.trim() ||
                `${getCalendarProviderName(connection)} could not be refreshed.`}{" "}
              {connection.lastSyncedAt
                ? `Showing the last successful sync from ${format(new Date(connection.lastSyncedAt), "MMM d 'at' h:mm a")}.`
                : "No successful calendar sync is available yet."}
            </Text>
          </Box>
          <Button
            className="text-base"
            color="tertiary"
            loading={isSyncing}
            onClick={() => {
              onSync(connection.id);
            }}
            variant="outline"
          >
            Retry
          </Button>
        </Flex>
      </Box>
    ) : null}

    {conflictCount > 0 ? (
      <Box className="border-danger/30 border-b px-5 py-3">
        <Text fontSize="md" fontWeight="medium">
          {conflictCount === 1
            ? "A scheduled block now overlaps a meeting"
            : `${conflictCount} scheduled blocks now overlap meetings`}
        </Text>
        <Text color="muted" fontSize="md">
          Open a red block to choose another time. FortyOne will not move locked
          work without your approval.
        </Text>
      </Box>
    ) : null}
  </>
);
