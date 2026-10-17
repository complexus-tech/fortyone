"use client";

import { useState } from "react";
import { Avatar, Box, Button, Flex, Input, Popover, Select, Text } from "ui";
import {
  CloseIcon,
  LockKeyholeIcon,
  PlusIcon,
  SearchIcon,
  ShareIcon,
  UserMultiple02Icon,
  WorkspaceIcon,
} from "icons";
import { useSession } from "@/lib/auth/client";
import { useMembers } from "@/lib/hooks/members";
import type {
  DocumentMember,
  DocumentVisibility,
  WorkspaceDocument,
} from "./types";
import { DocumentPublicLink } from "./document-public-link";
import { useUpdateDocumentAccess } from "./hooks";

const visibilityOptions: {
  description: string;
  icon: typeof UserMultiple02Icon;
  label: string;
  value: DocumentVisibility;
}[] = [
  {
    value: "workspace",
    label: "Members",
    description: "Members can edit. Guests can view.",
    icon: WorkspaceIcon,
  },
  {
    value: "restricted",
    label: "Invited people",
    description:
      "Set whether invited members can comment or edit. Guests can only view.",
    icon: UserMultiple02Icon,
  },
  {
    value: "private",
    label: "Only the owner",
    description: "Only the owner can access this document in the workspace.",
    icon: LockKeyholeIcon,
  },
];

export const DocumentAccessMenu = ({
  document,
}: {
  document: WorkspaceDocument;
}) => {
  const [open, setOpen] = useState(false);
  const { data: session } = useSession();
  const [visibility, setVisibility] = useState(document.visibility);
  const [members, setMembers] = useState(document.sharedWith);
  const [search, setSearch] = useState("");
  const {
    data: workspaceMembers = [],
    isPending: membersPending,
    isError: membersError,
  } = useMembers();
  const updateAccess = useUpdateDocumentAccess(document.id);
  const memberById = new Map(
    workspaceMembers.map((member) => [member.id, member]),
  );
  const selectedMemberIds = new Set(members.map((member) => member.userId));
  const owner = memberById.get(document.createdBy);
  const isOwner = session?.user.id === document.createdBy;
  const ownerName =
    owner?.fullName ||
    owner?.username ||
    (isOwner ? session.user.name : undefined) ||
    "Document owner";
  const ownerImage =
    owner?.avatarUrl ?? (isOwner ? session.user.image : undefined);
  const normalizedSearch = search.trim().toLowerCase();
  const availableMembers = workspaceMembers.filter(
    (member) =>
      member.id !== document.createdBy &&
      !selectedMemberIds.has(member.id) &&
      member.isActive &&
      !member.isSystem &&
      `${member.fullName} ${member.username} ${member.email}`
        .toLowerCase()
        .includes(normalizedSearch),
  );
  let emptyMembersMessage = "No matching members.";
  if (membersError) emptyMembersMessage = "Could not load members.";
  if (membersPending) emptyMembersMessage = "Loading members…";

  const save = () => {
    updateAccess.mutate(
      {
        visibility,
        members:
          visibility === "restricted"
            ? members.map((member) => ({
                ...member,
                role:
                  memberById.get(member.userId)?.role === "guest"
                    ? "viewer"
                    : member.role,
              }))
            : [],
      },
      {
        onSuccess: (response) => {
          if (!response.data) return;
          setVisibility(response.data.visibility);
          setMembers(response.data.sharedWith);
          setSearch("");
        },
      },
    );
  };

  const selectedOption = visibilityOptions.find(
    (option) => option.value === visibility,
  )!;
  const SelectedIcon = selectedOption.icon;
  const accessChanged =
    visibility !== document.visibility ||
    (visibility === "restricted" &&
      (members.length !== document.sharedWith.length ||
        members.some(
          (member) =>
            !document.sharedWith.some(
              (saved) =>
                saved.userId === member.userId && saved.role === member.role,
            ),
        )));

  const resetAccess = () => {
    setVisibility(document.visibility);
    setMembers(document.sharedWith);
    setSearch("");
  };

  return (
    <Popover
      onOpenChange={(nextOpen) => {
        setOpen(nextOpen);
        if (nextOpen) resetAccess();
      }}
      open={open}
    >
      <Popover.Trigger asChild>
        <Button
          aria-label="Share document"
          className="shrink-0 px-2 text-base sm:px-3"
          color="tertiary"
          leftIcon={<ShareIcon className="size-5" />}
          size="sm"
          variant="outline"
        >
          <span className="hidden sm:inline">Share</span>
        </Button>
      </Popover.Trigger>
      <Popover.Content
        align="end"
        aria-label="Share document"
        className="bg-surface-elevated dark:bg-surface-elevated/80 mr-0 max-h-[min(87dvh,52rem)] w-[29rem] max-w-[calc(100vw-1rem)] overflow-y-auto rounded-2xl py-0 text-base"
      >
        <Box className="border-border border-b px-5 py-4">
          <Flex align="center" justify="between">
            <Text as="h2" fontSize="lg" fontWeight="semibold">
              Share document
            </Text>
            <Button
              aria-label="Close sharing"
              asIcon
              color="tertiary"
              onClick={() => {
                setOpen(false);
              }}
              variant="naked"
            >
              <CloseIcon className="size-5" />
            </Button>
          </Flex>
          <Text className="mt-1 truncate" color="muted">
            {document.title || "Untitled document"}
          </Text>
        </Box>
        <Box className="px-5 py-4">
          <Flex align="center" gap={3} justify="between">
            <label
              className="shrink-0 font-medium"
              htmlFor="document-workspace-access"
            >
              Workspace access
            </label>
            <Select
              disabled={updateAccess.isPending}
              onValueChange={(value) => {
                setVisibility(value as DocumentVisibility);
              }}
              value={visibility}
            >
              <Select.Trigger
                className="ml-auto h-[2.1rem] min-w-0 text-base"
                id="document-workspace-access"
              >
                <Flex align="center" className="min-w-0" gap={2}>
                  <SelectedIcon className="size-5 shrink-0" />
                  <span className="truncate">{selectedOption.label}</span>
                </Flex>
              </Select.Trigger>
              <Select.Content>
                {visibilityOptions.map(({ icon: Icon, label, value }) => (
                  <Select.Option
                    className="text-base"
                    key={value}
                    value={value}
                  >
                    <Flex align="center" gap={2}>
                      <Icon className="size-5" />
                      <span>{label}</span>
                    </Flex>
                  </Select.Option>
                ))}
              </Select.Content>
            </Select>
          </Flex>
          <Text className="mt-2 leading-relaxed" color="muted">
            {selectedOption.description}
          </Text>
          <Flex align="center" className="mt-4" gap={3}>
            <Avatar name={ownerName} size="sm" src={ownerImage} />
            <Text className="min-w-0 flex-1 truncate" fontWeight="medium">
              {ownerName}
              {isOwner ? " (You)" : ""}
            </Text>
            <Text color="muted">Owner</Text>
          </Flex>
          {visibility === "restricted" ? (
            <Box className="mt-3">
              <Box className="space-y-2">
                {members.map((member) => {
                  const person = memberById.get(member.userId);
                  const name = person?.fullName || person?.username || "Member";
                  const isGuest = person?.role === "guest";
                  return (
                    <Flex align="center" gap={2} key={member.userId}>
                      <Avatar name={name} size="sm" src={person?.avatarUrl} />
                      <Text className="min-w-0 flex-1 truncate" title={name}>
                        {name}
                      </Text>
                      <Select
                        disabled={updateAccess.isPending || isGuest || !person}
                        onValueChange={(role: DocumentMember["role"]) => {
                          setMembers((current) =>
                            current.map((entry) =>
                              entry.userId === member.userId
                                ? { ...entry, role }
                                : entry,
                            ),
                          );
                        }}
                        value={isGuest ? "viewer" : member.role}
                      >
                        <Select.Trigger
                          aria-label={`Access for ${name}`}
                          className="h-[2.1rem] shrink-0 text-base"
                        >
                          <Select.Input />
                        </Select.Trigger>
                        <Select.Content>
                          <Select.Option className="text-base" value="viewer">
                            {person && !isGuest ? "Can comment" : "Can view"}
                          </Select.Option>
                          <Select.Option className="text-base" value="editor">
                            Can edit
                          </Select.Option>
                        </Select.Content>
                      </Select>
                      <Button
                        aria-label={`Remove access for ${name}`}
                        asIcon
                        color="tertiary"
                        disabled={updateAccess.isPending}
                        onClick={() => {
                          setMembers((current) =>
                            current.filter(
                              (entry) => entry.userId !== member.userId,
                            ),
                          );
                        }}
                        variant="naked"
                      >
                        <CloseIcon className="size-4" />
                      </Button>
                    </Flex>
                  );
                })}
              </Box>
              <Box className="mt-4">
                <Input
                  aria-label="Find people to invite"
                  className="text-base"
                  disabled={updateAccess.isPending}
                  leftIcon={<SearchIcon className="size-5" />}
                  onChange={(event) => {
                    setSearch(event.target.value);
                  }}
                  placeholder="Add people by name or email"
                  value={search}
                />
                {normalizedSearch || members.length === 0 ? (
                  <Box className="mt-2 max-h-48 space-y-1 overflow-y-auto">
                    {availableMembers.map((member) => {
                      const name = member.fullName || member.username;
                      return (
                        <Button
                          aria-label={`Invite ${name}`}
                          className="h-auto min-h-11 w-full py-2 text-base"
                          color="tertiary"
                          disabled={updateAccess.isPending}
                          key={member.id}
                          onClick={() => {
                            setMembers((current) => [
                              ...current,
                              {
                                userId: member.id,
                                role:
                                  member.role === "guest" ? "viewer" : "editor",
                              },
                            ]);
                            setSearch("");
                          }}
                          variant="naked"
                        >
                          <Avatar
                            name={name}
                            size="sm"
                            src={member.avatarUrl}
                          />
                          <span className="min-w-0 flex-1 truncate text-left">
                            {name}
                          </span>
                          <PlusIcon aria-hidden className="size-4" />
                        </Button>
                      );
                    })}
                    {availableMembers.length === 0 ? (
                      <Text className="px-1 py-2" color="muted" role="status">
                        {emptyMembersMessage}
                      </Text>
                    ) : null}
                  </Box>
                ) : null}
              </Box>
            </Box>
          ) : null}
          {accessChanged ? (
            <Box className="mt-4">
              <Text className="mb-3" color="muted" role="status">
                {updateAccess.isPending
                  ? "Saving access changes…"
                  : "Save to apply these access changes."}
              </Text>
              <Flex gap={2} justify="end">
                <Button
                  className="text-base"
                  color="tertiary"
                  disabled={updateAccess.isPending}
                  onClick={resetAccess}
                  variant="outline"
                >
                  Cancel
                </Button>
                <Button
                  className="text-base"
                  disabled={updateAccess.isPending}
                  onClick={save}
                >
                  {updateAccess.isPending ? "Saving…" : "Save changes"}
                </Button>
              </Flex>
            </Box>
          ) : null}
        </Box>
        <DocumentPublicLink document={document} />
      </Popover.Content>
    </Popover>
  );
};
