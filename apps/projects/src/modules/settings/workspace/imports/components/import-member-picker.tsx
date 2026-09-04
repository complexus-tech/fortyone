"use client";

import { useMemo, useState } from "react";
import { ArrowDown2Icon, CheckIcon } from "icons";
import { Avatar, Button, Command, Flex, Popover, Text } from "ui";
import type { Member } from "@/types/member";

type ImportMemberPickerProps = {
  disabled?: boolean;
  members: Member[];
  onChange: (memberId: string | null) => void;
  suggestedMemberId?: string;
  value: string | null;
};

const getMemberDisplayName = (member: Member) =>
  member.fullName || member.username;

export const ImportMemberPicker = ({
  disabled = false,
  members,
  onChange,
  suggestedMemberId,
  value,
}: ImportMemberPickerProps) => {
  const [open, setOpen] = useState(false);
  const eligibleMembers = useMemo(
    () =>
      members
        .filter((member) => member.isActive && !member.isSystem)
        .sort((left, right) => {
          if (left.id === suggestedMemberId) return -1;
          if (right.id === suggestedMemberId) return 1;
          return getMemberDisplayName(left).localeCompare(
            getMemberDisplayName(right),
          );
        }),
    [members, suggestedMemberId],
  );
  const selectedMember = eligibleMembers.find((member) => member.id === value);

  return (
    <Popover onOpenChange={setOpen} open={open}>
      <Popover.Trigger asChild>
        <Button
          align="between"
          aria-label={`Map source identity${selectedMember ? ` to ${getMemberDisplayName(selectedMember)}` : " to a workspace member"}`}
          className="max-w-full"
          color="tertiary"
          disabled={disabled}
          rightIcon={<ArrowDown2Icon className="h-4 shrink-0" />}
          size="sm"
          variant="outline"
        >
          <Text className="truncate">
            {selectedMember
              ? getMemberDisplayName(selectedMember)
              : "Choose member"}
          </Text>
        </Button>
      </Popover.Trigger>
      <Popover.Content
        align="end"
        className="z-[60] w-80 max-w-[calc(100vw-2rem)]"
      >
        <Command label="Workspace members">
          <Command.Input autoFocus placeholder="Search members…" />
          <Command.Separator />
          <Command.Empty className="py-3 text-base">
            <Text color="muted">No members found.</Text>
          </Command.Empty>
          <Command.Group className="max-h-64 overflow-y-auto">
            <Command.Item
              active={!value}
              className="justify-between opacity-70"
              onSelect={() => {
                onChange(null);
                setOpen(false);
              }}
              value="Unassigned"
            >
              <Flex align="center" gap={2}>
                <Avatar
                  className="text-foreground/80"
                  color="primary"
                  size="sm"
                />
                <Text className="max-w-40 truncate">Unassigned</Text>
              </Flex>
              {!value ? (
                <CheckIcon
                  className="text-foreground h-5 w-auto"
                  strokeWidth={2.1}
                />
              ) : null}
            </Command.Item>
            {eligibleMembers.map((member) => (
              <Command.Item
                active={value === member.id}
                className="justify-between"
                key={member.id}
                onSelect={() => {
                  onChange(member.id);
                  setOpen(false);
                }}
                value={`${getMemberDisplayName(member)} ${member.email}`}
              >
                <Flex align="center" className="min-w-0 flex-1" gap={2}>
                  <Avatar
                    aria-hidden="true"
                    color="primary"
                    name={getMemberDisplayName(member)}
                    size="sm"
                    src={member.avatarUrl}
                  />
                  <Text className="min-w-0 flex-1 truncate font-medium">
                    {getMemberDisplayName(member)}
                  </Text>
                </Flex>
                {value === member.id ? (
                  <CheckIcon className="h-4 shrink-0" strokeWidth={2.2} />
                ) : null}
              </Command.Item>
            ))}
          </Command.Group>
        </Command>
      </Popover.Content>
    </Popover>
  );
};
