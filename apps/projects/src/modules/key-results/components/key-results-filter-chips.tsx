import type { ReactNode } from "react";
import { CloseIcon } from "icons";
import { Avatar, Flex, Popover } from "ui";

type KeyResultsLeadChipMember = {
  avatarUrl?: string | null;
  id: string;
  name: string;
  username: string;
};

export const KeyResultsLeadChipValue = ({
  members,
}: {
  members: KeyResultsLeadChipMember[];
}) => {
  const visibleMembers = members.slice(0, 2);

  if (members.length > 2) {
    return (
      <Flex align="center" gap={1}>
        <Flex align="center" className="-space-x-1">
          {visibleMembers.map((member) => (
            <Avatar
              className="ring-background ring-1"
              color="primary"
              key={member.id}
              name={member.name}
              size="xs"
              src={member.avatarUrl}
            />
          ))}
        </Flex>
        <span>{members.length} leads</span>
      </Flex>
    );
  }

  return (
    <Flex align="center" gap={2}>
      {visibleMembers.map((member) => (
        <Flex align="center" gap={1} key={member.id}>
          <Avatar
            color="primary"
            name={member.name}
            size="xs"
            src={member.avatarUrl}
          />
          <span>{member.username}</span>
        </Flex>
      ))}
    </Flex>
  );
};

export const KeyResultsFilterChip = ({
  editor,
  icon,
  label,
  onRemove,
  operator,
  value,
}: {
  editor: ReactNode;
  icon: ReactNode;
  label: string;
  onRemove: () => void;
  operator: string;
  value: ReactNode;
}) => (
  <Flex
    align="center"
    className="border-border bg-surface h-[2.1rem] shrink-0 overflow-hidden rounded-xl border"
    gap={0}
  >
    <span className="border-border text-text-secondary flex h-full items-center gap-1.5 border-r px-2.5">
      {icon}
      {label}
    </span>
    <span className="border-border text-text-secondary flex h-full items-center border-r px-2.5">
      {operator}
    </span>
    <Popover>
      <Popover.Trigger asChild>
        <button
          className="hover:bg-state-hover flex h-full max-w-72 items-center truncate px-2.5 text-left transition"
          type="button"
        >
          <span className="flex min-w-0 items-center truncate">{value}</span>
        </button>
      </Popover.Trigger>
      <Popover.Content align="start" className="w-80 p-4">
        {editor}
      </Popover.Content>
    </Popover>
    <button
      aria-label={`Remove ${label} filter`}
      className="hover:bg-state-hover border-border flex h-full w-9 items-center justify-center border-l transition"
      onClick={onRemove}
      type="button"
    >
      <CloseIcon className="text-text-secondary h-3.5 w-auto" />
    </button>
  </Flex>
);
