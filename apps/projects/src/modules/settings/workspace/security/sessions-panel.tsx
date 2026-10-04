"use client";

import { useRef, useState } from "react";
import { Box, Button, Dialog, Flex, Select, Text } from "ui";
import { useMembers } from "@/lib/hooks/members";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { useRevokeSessions, useSecuritySessions } from "./hooks";
import { SecurityInput } from "./security-input";
import { SessionList } from "./session-list";
import { SessionBrowser } from "./session-browser";
import { SecurityTimestamp } from "./security-timestamp";
import { sessionStatus } from "./types";

type Target = {
  id: string;
  name: string;
  member: boolean;
  current: boolean;
  browserName?: string | null;
  lastSeenAt?: string;
};
export const SecuritySessions = () => {
  const [userId, setUserId] = useState("");
  const [target, setTarget] = useState<Target | null>(null);
  const [reason, setReason] = useState("");
  const returnFocusRef = useRef<HTMLButtonElement | null>(null);
  const memberFilterRef = useRef<HTMLButtonElement>(null);
  const { data: members = [] } = useMembers();
  const query = useSecuritySessions(userId, false);
  const sessions =
    query.data?.items.filter(
      (session) => sessionStatus(session) === "Active",
    ) ?? [];
  const usernames = new Map(
    members.map((member) => [member.id, member.username]),
  );
  const revoke = useRevokeSessions();
  const choose = (next: Target, trigger: HTMLButtonElement | null) => {
    if (revoke.isPending) return;
    returnFocusRef.current = trigger;
    revoke.reset();
    setReason("");
    setTarget(next);
  };
  const selectedMember = members.find((member) => member.id === userId);
  const close = () => {
    if (!revoke.isPending) setTarget(null);
  };
  let description = `End this session's access to the workspace for ${target?.name ?? "this member"}.`;
  if (target?.member) {
    description = `Require ${target.name} to log in again before accessing this workspace.`;
  } else if (target?.current) {
    description =
      "Your current session will lose access to this workspace. Log in again to return.";
  }
  return (
    <>
      <Box className="border-border bg-surface overflow-hidden rounded-2xl border">
        <SectionHeader
          action={
            <Button
              color="tertiary"
              disabled={query.isFetching}
              loading={Boolean(query.isFetching && !query.isPending)}
              loadingText="Refreshing…"
              onClick={() => void query.refetch()}
              variant="outline"
            >
              Refresh sessions
            </Button>
          }
          description="Active browser sessions for workspace members. All times are shown in UTC."
          title="Member sessions"
        />
        <Flex align="end" className="gap-4 px-6 py-4" wrap>
          <Box className="w-full min-w-0 md:w-auto md:max-w-64">
            <label
              className="mb-[0.35rem] block"
              htmlFor="security-session-member"
            >
              Member
            </label>
            <Select
              onValueChange={(value) => {
                setUserId(value === "all" ? "" : value);
              }}
              value={userId || "all"}
            >
              <Select.Trigger
                className="text-base"
                id="security-session-member"
                ref={memberFilterRef}
              >
                <Select.Input />
              </Select.Trigger>
              <Select.Content>
                <Select.Option className="text-base" value="all">
                  All members
                </Select.Option>
                {members
                  .filter((member) => !member.isSystem)
                  .map((member) => (
                    <Select.Option
                      className="text-base"
                      key={member.id}
                      value={member.id}
                    >
                      {member.username}
                    </Select.Option>
                  ))}
              </Select.Content>
            </Select>
          </Box>
          {selectedMember ? (
            <Button
              color="danger"
              onClick={(event) => {
                choose(
                  {
                    id: selectedMember.id,
                    name: selectedMember.username,
                    member: true,
                    current: false,
                  },
                  event.currentTarget,
                );
              }}
              variant="outline"
            >
              Revoke all member sessions
            </Button>
          ) : null}
        </Flex>
        {query.isPending ? (
          <Text aria-live="polite" className="px-6 py-4" color="muted">
            Loading member sessions…
          </Text>
        ) : null}
        {query.isError ? (
          <Box className="px-6 py-4">
            <Text color="danger" role="alert">
              Member sessions could not be loaded.
            </Text>
            <Button
              className="mt-3"
              color="tertiary"
              onClick={() => void query.refetch()}
              variant="outline"
            >
              Try again
            </Button>
          </Box>
        ) : null}
        {!query.isPending && !query.isError && sessions.length === 0 ? (
          <Text className="px-6 py-4" color="muted">
            No active sessions match this member filter.
          </Text>
        ) : null}
        {!query.isError && sessions.length ? (
          <SessionList
            onRevoke={(session, trigger) => {
              choose(
                {
                  id: session.id,
                  name:
                    session.username ||
                    usernames.get(session.userId) ||
                    session.name,
                  member: false,
                  current: session.current,
                  browserName: session.browserName,
                  lastSeenAt: session.lastSeenAt,
                },
                trigger,
              );
            }}
            sessions={sessions}
            usernames={usernames}
          />
        ) : null}
        {!query.isError && query.data?.hasMore ? (
          <Text className="px-6 py-4" color="muted">
            Showing the 500 most recently used sessions. Select a member to
            narrow the results.
          </Text>
        ) : null}
      </Box>
      <Dialog
        onOpenChange={(open) => {
          if (!open) close();
        }}
        open={Boolean(target)}
      >
        <Dialog.Content
          aria-busy={revoke.isPending}
          className="flex max-h-[calc(100dvh-15vw-1rem)] flex-col md:max-h-[calc(100dvh-10vw-1rem)]"
          hideClose={revoke.isPending}
          onCloseAutoFocus={(event) => {
            const trigger = returnFocusRef.current?.isConnected
              ? returnFocusRef.current
              : memberFilterRef.current;
            if (trigger?.isConnected) {
              event.preventDefault();
              trigger.focus();
            }
            returnFocusRef.current = null;
          }}
          onEscapeKeyDown={(event) => {
            if (revoke.isPending) event.preventDefault();
          }}
          onInteractOutside={(event) => {
            if (revoke.isPending) event.preventDefault();
          }}
        >
          <Dialog.Header className="shrink-0 px-6 py-5">
            <Dialog.Title className="text-lg">
              {target?.member ? "Revoke member sessions" : "Revoke session"}
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="max-h-none min-h-0 flex-1 space-y-4">
            <Dialog.Description asChild className="px-0 text-base">
              <Text color="muted">{description}</Text>
            </Dialog.Description>
            {target && !target.member ? (
              <Flex align="center" className="gap-x-4 gap-y-2" wrap>
                <SessionBrowser
                  current={target.current}
                  name={target.browserName}
                />
                {target.lastSeenAt ? (
                  <Text color="muted">
                    Last active <SecurityTimestamp value={target.lastSeenAt} />
                  </Text>
                ) : null}
              </Flex>
            ) : null}
            <SecurityInput
              autoFocus
              disabled={revoke.isPending}
              label="Reason for revocation"
              maxLength={240}
              onChange={(event) => {
                setReason(event.target.value);
              }}
              placeholder="Lost device or access review"
              value={reason}
            />
            <Text color="muted">
              The reason is recorded in the workspace audit log.
            </Text>
            {revoke.isError ? (
              <Text color="danger" role="alert">
                {revoke.error.message}
              </Text>
            ) : null}
          </Dialog.Body>
          <Dialog.Footer className="shrink-0 gap-3 py-4" justify="end">
            <Button
              color="tertiary"
              disabled={revoke.isPending}
              onClick={close}
              variant="outline"
            >
              Cancel
            </Button>
            <Button
              color="danger"
              disabled={!reason.trim() || revoke.isPending}
              loading={revoke.isPending}
              loadingText="Revoking…"
              onClick={() => {
                if (!target || revoke.isPending || !reason.trim()) return;
                revoke.mutate(
                  {
                    id: target.id,
                    member: target.member,
                    reason: reason.trim(),
                  },
                  {
                    onSuccess: () => {
                      setTarget(null);
                    },
                  },
                );
              }}
            >
              Revoke access
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
