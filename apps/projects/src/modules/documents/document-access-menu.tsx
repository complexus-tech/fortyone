"use client";

import { useState } from "react";
import {
  Avatar,
  Box,
  Button,
  Command,
  Divider,
  Flex,
  Popover,
  Select,
  Text,
} from "ui";
import {
  CheckIcon,
  CloseIcon,
  LockKeyholeIcon,
  ShareIcon,
  UserMultiple02Icon,
  WorkspaceIcon,
} from "icons";
import { useSession } from "@/lib/auth/client";
import { useMembers } from "@/lib/hooks/members";
import { DocumentPublicLink } from "./document-public-link";
import { useUpdateDocumentAccess } from "./hooks";
import type {
  DocumentMember,
  DocumentVisibility,
  WorkspaceDocument,
} from "./types";

const visibilityOptions: {
  description: string;
  icon: typeof UserMultiple02Icon;
  label: string;
  value: DocumentVisibility;
}[] = [
  {
    value: "workspace",
    label: "Workspace",
    description: "Workspace members can edit. Guests can view.",
    icon: WorkspaceIcon,
  },
  {
    value: "restricted",
    label: "Selected people",
    description: "Only you and selected people.",
    icon: UserMultiple02Icon,
  },
  {
    value: "private",
    label: "Only me",
    description: "Only you can access it in the workspace.",
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
  const [visibility, setVisibility] = useState<DocumentVisibility>("workspace");
  const [members, setMembers] = useState<DocumentMember[]>([]);
  const [search, setSearch] = useState("");
  const { data: workspaceMembers = [] } = useMembers(search);
  const updateAccess = useUpdateDocumentAccess(document.id);
  const selectedMemberIds = new Set(members.map((member) => member.userId));
  const availableMembers = workspaceMembers.filter(
    (member) => member.id !== document.createdBy,
  );

  const toggleMember = (userId: string, role: DocumentMember["role"]) => {
    setMembers((current) =>
      current.some((member) => member.userId === userId)
        ? current.filter((member) => member.userId !== userId)
        : [...current, { userId, role }],
    );
  };

  const save = () => {
    updateAccess.mutate(
      {
        visibility,
        members: visibility === "restricted" ? members : [],
      },
      {
        onSuccess: () => {
          setOpen(false);
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

  const handleOpenChange = (nextOpen: boolean) => {
    setOpen(nextOpen);
    if (!nextOpen) return;
    setVisibility(document.visibility);
    setMembers(document.sharedWith);
    setSearch("");
  };

  return (
    <Popover onOpenChange={handleOpenChange} open={open}>
      <Popover.Trigger asChild>
        <Button
          aria-label="Share document"
          color="tertiary"
          leftIcon={<ShareIcon className="size-4" />}
          size="sm"
          variant="outline"
        >
          Share
        </Button>
      </Popover.Trigger>
      <Popover.Content
        align="end"
        className="bg-surface-elevated dark:bg-surface-elevated/80 mr-0 max-h-[87vh] w-[26rem] max-w-[calc(100vw-1rem)] overflow-y-auto rounded-2xl py-0"
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
              size="sm"
              variant="naked"
            >
              <CloseIcon className="size-5" />
            </Button>
          </Flex>
          <Text className="mt-1 truncate" color="muted">
            {document.title}
          </Text>
        </Box>
        <Box className="px-5 pt-4 pb-3">
          <label
            className="mb-2 block font-medium"
            htmlFor="document-workspace-access"
          >
            Who has access
          </label>
          <Select
            disabled={updateAccess.isPending}
            onValueChange={(value) => {
              setVisibility(value as DocumentVisibility);
            }}
            value={visibility}
          >
            <Select.Trigger
              className="h-11 text-base"
              id="document-workspace-access"
            >
              <Flex align="center" gap={2}>
                <SelectedIcon className="size-4" />
                <span>{selectedOption.label}</span>
              </Flex>
            </Select.Trigger>
            <Select.Content>
              {visibilityOptions.map(({ icon: Icon, label, value }) => (
                <Select.Option key={value} value={value}>
                  <Flex align="center" gap={2}>
                    <Icon className="size-4" />
                    <span>{label}</span>
                  </Flex>
                </Select.Option>
              ))}
            </Select.Content>
          </Select>
          <Text className="mt-2 leading-relaxed" color="muted">
            {selectedOption.description}
          </Text>
        </Box>
        <Flex align="center" className="px-5 pt-1 pb-4" gap={3}>
          <Avatar
            name={session?.user.name || "Document owner"}
            size="sm"
            src={session?.user.image}
          />
          <Box className="min-w-0 flex-1">
            <Text className="truncate" fontWeight="medium">
              {session?.user.name || "You"}
            </Text>
            <Text className="truncate" color="muted">
              {session?.user.email}
            </Text>
          </Box>
          <Text color="muted">Owner</Text>
        </Flex>
        {visibility === "restricted" ? (
          <Box className="px-3 pb-4">
            <Divider className="mb-3" />
            <Box className="px-2">
              <Command shouldFilter={false}>
                <Command.Input
                  autoFocus
                  onValueChange={setSearch}
                  placeholder="Search workspace members..."
                  value={search}
                />
                <Divider className="my-2" />
              </Command>
              <Box className="mt-2 max-h-52 overflow-y-auto">
                {availableMembers.map((member) => {
                  const selected = selectedMemberIds.has(member.id);
                  return (
                    <button
                      aria-pressed={selected}
                      className="hover:bg-state-hover flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left"
                      key={member.id}
                      onClick={() => {
                        toggleMember(
                          member.id,
                          member.role === "guest" ? "viewer" : "editor",
                        );
                      }}
                      type="button"
                    >
                      <Avatar
                        name={member.fullName || member.username}
                        size="xs"
                        src={member.avatarUrl}
                      />
                      <Text className="min-w-0 flex-1 truncate">
                        {member.fullName || member.username}
                      </Text>
                      <Text color="muted">
                        {member.role === "guest" ? "Can view" : "Can edit"}
                      </Text>
                      {selected ? (
                        <CheckIcon className="text-primary size-4" />
                      ) : null}
                    </button>
                  );
                })}
              </Box>
            </Box>
          </Box>
        ) : null}

        {accessChanged ? (
          <Flex className="px-5 pb-4" gap={2} justify="end">
            <Button
              color="tertiary"
              disabled={updateAccess.isPending}
              onClick={() => {
                setVisibility(document.visibility);
                setMembers(document.sharedWith);
              }}
              size="sm"
              variant="outline"
            >
              Cancel
            </Button>
            <Button disabled={updateAccess.isPending} onClick={save} size="sm">
              {updateAccess.isPending ? "Saving…" : "Save changes"}
            </Button>
          </Flex>
        ) : null}
        <DocumentPublicLink document={document} />
      </Popover.Content>
    </Popover>
  );
};
