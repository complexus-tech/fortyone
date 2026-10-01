"use client";

import { useState } from "react";
import {
  Badge,
  Box,
  Button,
  Checkbox,
  Dialog,
  Flex,
  Input,
  Select,
  Text,
} from "ui";
import { useMembers } from "@/lib/hooks/members";
import { SectionHeader } from "@/modules/settings/components/section-header";
import type { BrowserSession } from "./types";
import { useRevokeSessions, useSecuritySessions } from "./hooks";
import { sessionStatus } from "./types";

type Target = { id: string; name: string; member: boolean; current: boolean };
const SessionRow = ({
  session,
  onRevoke,
}: {
  session: BrowserSession;
  onRevoke: () => void;
}) => {
  const status = sessionStatus(session);
  return (
    <Flex align="start" className="gap-5 p-5" justify="between" wrap>
      <Box className="min-w-0 flex-1">
        <Flex align="center" className="gap-2" wrap>
          <Text fontWeight="medium">{session.name}</Text>
          <Badge color="tertiary" variant="outline">
            {status}
          </Badge>
          {session.current ? (
            <Badge color="tertiary" variant="outline">
              Current session
            </Badge>
          ) : null}
        </Flex>
        <Text className="mt-1 break-all" color="muted">
          {session.email} · {session.role}
        </Text>
        <Box className="mt-3 grid gap-2 xl:grid-cols-3">
          <Text color="muted">
            Logged in {new Date(session.authenticatedAt).toLocaleString()}
          </Text>
          <Text color="muted">
            Last seen {new Date(session.lastSeenAt).toLocaleString()}
          </Text>
          <Text color="muted">
            Expires {new Date(session.expiresAt).toLocaleString()}
          </Text>
        </Box>
      </Box>
      {status === "Active" ? (
        <Button color="danger" onClick={onRevoke} variant="outline">
          Revoke session
        </Button>
      ) : null}
    </Flex>
  );
};
export const SecuritySessions = () => {
  const [userId, setUserId] = useState("");
  const [includeRevoked, setIncludeRevoked] = useState(false);
  const [target, setTarget] = useState<Target | null>(null);
  const [reason, setReason] = useState("");
  const { data: members = [] } = useMembers();
  const query = useSecuritySessions(userId, includeRevoked);
  const revoke = useRevokeSessions();
  const choose = (next: Target) => {
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
          description="Review and revoke access to this workspace."
          title="Member sessions"
        />
        <Flex align="end" className="gap-4 px-6 pt-5 pb-5" wrap>
          <Box className="w-full max-w-sm">
            <label className="mb-2 block" htmlFor="security-session-member">
              Member
            </label>
            <Select
              onValueChange={(value) => {
                setUserId(value === "all" ? "" : value);
              }}
              value={userId || "all"}
            >
              <Select.Trigger
                className="h-11 text-base"
                id="security-session-member"
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
                      {member.fullName || member.email}
                    </Select.Option>
                  ))}
              </Select.Content>
            </Select>
          </Box>
          <label className="flex h-11 cursor-pointer items-center gap-3">
            <Checkbox
              checked={includeRevoked}
              onCheckedChange={(checked) => {
                setIncludeRevoked(checked === true);
              }}
            />{" "}
            <Text>Include revoked and expired sessions</Text>
          </label>
          {selectedMember ? (
            <Button
              color="danger"
              onClick={() => {
                choose({
                  id: selectedMember.id,
                  name: selectedMember.fullName || selectedMember.email,
                  member: true,
                  current: false,
                });
              }}
              variant="outline"
            >
              Revoke all member sessions
            </Button>
          ) : null}
        </Flex>
        {query.isPending ? (
          <Text aria-live="polite" className="px-6 pt-5 pb-6" color="muted">
            Loading member sessions…
          </Text>
        ) : null}
        {query.isError ? (
          <Box className="px-6 pt-5 pb-6">
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
        {query.data?.items.length === 0 ? (
          <Text className="px-6 pt-5 pb-6" color="muted">
            No recorded sessions match these filters.
          </Text>
        ) : null}
        <Box className="divide-border border-border divide-y border-t">
          {query.data?.items.map((session) => (
            <SessionRow
              key={session.id}
              onRevoke={() => {
                choose({
                  id: session.id,
                  name: session.name,
                  member: false,
                  current: session.current,
                });
              }}
              session={session}
            />
          ))}
        </Box>
        {query.data?.hasMore ? (
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
          className="flex max-h-[calc(100dvh-15vw-1rem)] flex-col md:max-h-[calc(100dvh-10vw-1rem)]"
          hideClose={revoke.isPending}
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
            <Dialog.Description asChild className="px-0 text-base leading-6">
              <Text color="muted">{description}</Text>
            </Dialog.Description>
            <Input
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
              onClick={() => {
                if (target)
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
              {revoke.isPending ? "Revoking…" : "Revoke access"}
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
