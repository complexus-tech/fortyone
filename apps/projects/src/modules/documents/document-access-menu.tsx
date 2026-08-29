"use client";

import { useState } from "react";
import { Avatar, Box, Button, Command, Divider, Flex, Popover, Text } from "ui";
import {
  CheckIcon,
  LockKeyholeIcon,
  ShareIcon,
  UserMultiple02Icon,
} from "icons";
import { cn } from "lib";
import { useMembers } from "@/lib/hooks/members";
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
    label: "Everyone in the workspace",
    description: "Everyone in the workspace can find and access this document.",
    icon: UserMultiple02Icon,
  },
  {
    value: "restricted",
    label: "Selected people",
    description: "Only you and invited people can access this document.",
    icon: UserMultiple02Icon,
  },
  {
    value: "private",
    label: "Only me",
    description: "This document is private to its creator.",
    icon: LockKeyholeIcon,
  },
];

export const DocumentAccessMenu = ({
  document,
}: {
  document: WorkspaceDocument;
}) => {
  const [open, setOpen] = useState(false);
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

  const activeOption = visibilityOptions.find(
    (option) => option.value === document.visibility,
  );
  const ActiveIcon = activeOption?.icon ?? ShareIcon;

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
          aria-label={`Document access: ${activeOption?.label ?? "Share"}`}
          asIcon
          color="tertiary"
          size="sm"
          variant="outline"
        >
          <ActiveIcon className="size-4" />
        </Button>
      </Popover.Trigger>
      <Popover.Content
        align="end"
        className="bg-surface-elevated dark:bg-surface-elevated/80 mr-0 max-h-[87vh] w-[26rem] max-w-[calc(100vw-1rem)] overflow-y-auto rounded-2xl pb-2"
      >
        <Flex align="center" className="h-11 px-4">
          <Text fontWeight="semibold">Document access</Text>
        </Flex>
        <Divider className="mb-1.5" />
        <Box className="space-y-1 px-2">
          {visibilityOptions.map(
            ({ description, icon: Icon, label, value }) => (
              <button
                className={cn(
                  "hover:bg-state-hover flex w-full items-start gap-3 rounded-xl px-2.5 py-2.5 text-left",
                  { "bg-state-active": visibility === value },
                )}
                key={value}
                onClick={() => {
                  setVisibility(value);
                }}
                type="button"
              >
                <Icon className="text-text-muted mt-0.5 size-5 shrink-0" />
                <span className="min-w-0 flex-1">
                  <Text fontWeight="medium">{label}</Text>
                  <Text color="muted">{description}</Text>
                </span>
                {visibility === value ? (
                  <CheckIcon className="text-primary mt-0.5 size-5" />
                ) : null}
              </button>
            ),
          )}
        </Box>

        {visibility === "restricted" ? (
          <Box className="mt-3">
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

        <Divider className="mt-3" />
        <Flex className="px-4 pt-2" justify="end">
          <Button
            color="primary"
            disabled={updateAccess.isPending}
            onClick={save}
            size="sm"
          >
            {updateAccess.isPending ? "Saving..." : "Save access"}
          </Button>
        </Flex>
      </Popover.Content>
    </Popover>
  );
};
