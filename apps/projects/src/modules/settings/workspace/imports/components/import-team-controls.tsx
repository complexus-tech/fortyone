"use client";
import type { ReactNode } from "react";
import { useState } from "react";
import { ArrowDown2Icon, CheckIcon } from "icons";
import { cn } from "lib";
import { Box, Button, Command, Flex, Popover, Text } from "ui";
import { TeamColor } from "@/components/ui/team-color";
import type { Team } from "@/modules/teams/public/types";

export const SelectionCard = ({
  description,
  disabled = false,
  icon,
  label,
  onClick,
  selected,
}: {
  description: string;
  disabled?: boolean;
  icon: ReactNode;
  label: string;
  onClick: () => void;
  selected: boolean;
}) => (
  <button
    aria-pressed={selected}
    className={cn(
      "border-border bg-surface hover:border-border-strong focus-visible:ring-ring relative min-h-20 rounded-xl border-[0.5px] p-3 text-left transition-[border-color,background-color,box-shadow] focus-visible:ring-2 focus-visible:outline-none",
      selected && "border-primary bg-primary/5 ring-primary/15 ring-1",
      disabled && "cursor-not-allowed opacity-60",
    )}
    disabled={disabled}
    onClick={onClick}
    type="button"
  >
    <Flex align="start" gap={3}>
      <Box
        className={cn(
          "bg-surface-muted text-text-muted flex h-9 w-9 shrink-0 items-center justify-center rounded-lg transition-colors",
          selected && "bg-primary/10 text-primary",
        )}
      >
        {icon}
      </Box>
      <Box className="min-w-0">
        <Text className="font-medium">{label}</Text>
        <Text className="mt-1 leading-5" color="muted">
          {description}
        </Text>
      </Box>
    </Flex>
  </button>
);

export const DestinationTeamPicker = ({
  disabled = false,
  onChange,
  teams,
  value,
}: {
  disabled?: boolean;
  onChange: (teamId: string) => void;
  teams: Team[];
  value: string;
}) => {
  const [open, setOpen] = useState(false);
  const selectedTeam = teams.find((team) => team.id === value);

  if (teams.length === 0) {
    return (
      <Text className="mt-3" color="muted">
        No existing teams. Choose “Create a new team” to continue.
      </Text>
    );
  }

  return (
    <Popover onOpenChange={setOpen} open={open}>
      <Popover.Trigger asChild>
        <Button
          align="between"
          className="mt-3 max-w-full"
          color="tertiary"
          disabled={disabled}
          rightIcon={<ArrowDown2Icon className="h-4 shrink-0" />}
          size="sm"
          variant="outline"
        >
          <Flex align="center" className="min-w-0" gap={2}>
            <TeamColor className="shrink-0" color={selectedTeam?.color} />
            <Text className="truncate">
              {selectedTeam?.name ?? "Choose a team"}
            </Text>
            {selectedTeam ? (
              <Text className="shrink-0" color="muted">
                {selectedTeam.code}
              </Text>
            ) : null}
          </Flex>
        </Button>
      </Popover.Trigger>
      <Popover.Content
        align="start"
        className="z-[60] w-80 max-w-[calc(100vw-2rem)]"
      >
        <Command label="Destination teams">
          <Command.Input autoFocus placeholder="Search teams…" />
          <Command.Separator />
          <Command.Empty className="py-3 text-base">
            <Text color="muted">No teams found.</Text>
          </Command.Empty>
          <Command.Group className="max-h-60 overflow-y-auto">
            {teams.map((team) => (
              <Command.Item
                active={value === team.id}
                className="justify-between"
                key={team.id}
                onSelect={() => {
                  onChange(team.id);
                  setOpen(false);
                }}
              >
                <Flex align="center" className="min-w-0" gap={2}>
                  <TeamColor className="shrink-0" color={team.color} />
                  <Text className="truncate">{team.name}</Text>
                </Flex>
                <Flex align="center" className="shrink-0" gap={2}>
                  <Text color="muted">{team.code}</Text>
                  {value === team.id ? (
                    <CheckIcon className="h-4 w-auto" strokeWidth={2.2} />
                  ) : null}
                </Flex>
              </Command.Item>
            ))}
          </Command.Group>
        </Command>
      </Popover.Content>
    </Popover>
  );
};
