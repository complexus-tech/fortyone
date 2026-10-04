import { useRef } from "react";
import { LogoutIcon, MoreHorizontalIcon } from "icons";
import { Box, Button, Menu, Table, Text } from "ui";
import type { BrowserSession } from "./types";
import { sessionStatus } from "./types";
import { SessionBrowser } from "./session-browser";
import { SecurityTimestamp } from "./security-timestamp";

type RevokeSession = (
  session: BrowserSession,
  trigger: HTMLButtonElement | null,
) => void;

const SessionActions = ({
  session,
  username,
  onRevoke,
}: {
  session: BrowserSession;
  username: string;
  onRevoke: RevokeSession;
}) => {
  const triggerRef = useRef<HTMLButtonElement>(null);
  const openingConfirmation = useRef(false);
  return (
    <Menu>
      <Menu.Button>
        <Button
          aria-label={`Session actions for ${username} (${session.id.slice(0, 8)})`}
          asIcon
          color="tertiary"
          leftIcon={<MoreHorizontalIcon aria-hidden />}
          ref={triggerRef}
          size="sm"
          variant="naked"
        />
      </Menu.Button>
      <Menu.Items
        align="end"
        onCloseAutoFocus={(event) => {
          if (!openingConfirmation.current) return;
          event.preventDefault();
          openingConfirmation.current = false;
        }}
      >
        <Menu.Group>
          <Menu.Item
            className="text-danger"
            disabled={sessionStatus(session) !== "Active"}
            onSelect={() => {
              if (sessionStatus(session) !== "Active") return;
              openingConfirmation.current = true;
              onRevoke(session, triggerRef.current);
            }}
          >
            <LogoutIcon aria-hidden className="h-[1.15rem]" />
            Revoke session
          </Menu.Item>
        </Menu.Group>
      </Menu.Items>
    </Menu>
  );
};

export const SessionList = ({
  sessions,
  usernames,
  onRevoke,
}: {
  sessions: BrowserSession[];
  usernames: ReadonlyMap<string, string>;
  onRevoke: RevokeSession;
}) => {
  return (
    <Box
      aria-label="Member sessions table"
      className="border-border focus-visible:ring-ring overflow-x-auto border-t-[0.5px] focus-visible:ring-2 focus-visible:outline-none focus-visible:ring-inset"
      role="region"
      tabIndex={0}
    >
      <Table
        align={null}
        className="min-w-[58rem] text-base [&_td]:px-6 [&_td]:py-3 [&_td]:whitespace-nowrap [&_td:first-child]:font-normal [&_td:last-child]:pr-6 [&_tr]:border-b-[0.5px]"
        color="light"
        size="md"
        variant="bordered"
      >
        <caption className="sr-only">
          Member browser sessions. All times are shown in UTC.
        </caption>
        <Table.Head>
          <Table.Tr>
            {[
              "Member",
              "Browser",
              "Last active",
              "Signed in",
              "Expires",
              "Actions",
            ].map((label) => (
              <Table.Th
                className="bg-surface static px-6 text-left text-base font-medium whitespace-nowrap normal-case"
                key={label}
                scope="col"
              >
                {label === "Actions" ? (
                  <span className="sr-only">Actions</span>
                ) : (
                  label
                )}
              </Table.Th>
            ))}
          </Table.Tr>
        </Table.Head>
        <Table.Body>
          {sessions.map((session) => {
            const username =
              session.username ||
              usernames.get(session.userId) ||
              "Unknown member";
            return (
              <Table.Tr key={session.id}>
                <Table.Td className="max-w-56">
                  <Text className="truncate" title={username}>
                    {username}
                  </Text>
                </Table.Td>
                <Table.Td>
                  <SessionBrowser
                    current={session.current}
                    name={session.browserName}
                  />
                </Table.Td>
                {[
                  ["lastActive", session.lastSeenAt],
                  ["signedIn", session.authenticatedAt],
                  ["expires", session.expiresAt],
                ].map(([field, value]) => (
                  <Table.Td key={field}>
                    <SecurityTimestamp value={value} />
                  </Table.Td>
                ))}
                <Table.Td className="w-12">
                  <SessionActions
                    onRevoke={onRevoke}
                    session={session}
                    username={username}
                  />
                </Table.Td>
              </Table.Tr>
            );
          })}
        </Table.Body>
      </Table>
    </Box>
  );
};
