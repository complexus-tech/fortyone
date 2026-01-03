"use client";

import { Command, Flex, Popover, Text, Divider, Button, Box } from "ui";
import type { ReactNode } from "react";
import { createContext, useContext, useState } from "react";
import { TeamIcon, PlusIcon } from "icons";
import { useRouter } from "next/navigation";
import { useTeams, usePublicTeams } from "@/modules/teams/hooks/teams";

type TeamContextType = {
  open: boolean;
  setOpen: (open: boolean) => void;
};

const TeamsContext = createContext<TeamContextType>({
  open: false,
  setOpen: () => {},
});

export const useTeamsMenu = () => {
  const context = useContext(TeamsContext);
  return context;
};

const Menu = ({ children }: { children: ReactNode }) => {
  const { open, setOpen } = useTeamsMenu();
  return (
    <Popover onOpenChange={setOpen} open={open}>
      {children}
    </Popover>
  );
};

export const TeamsMenu = ({ children }: { children: ReactNode }) => {
  const [open, setOpen] = useState(false);
  return (
    <TeamsContext.Provider value={{ open, setOpen }}>
      <Menu>{children}</Menu>
    </TeamsContext.Provider>
  );
};

const Items = ({
  hideManageTeams,
  setTeam,
}: {
  hideManageTeams?: boolean;
  setTeam: (teamId: string, action: "join" | "leave") => void;
}) => {
  const router = useRouter();
  const { data: teams = [] } = useTeams();
  const { data: publicTeams = [] } = usePublicTeams();
  const { setOpen } = useTeamsMenu();

  return (
    <Popover.Content align="start" className="w-72" sideOffset={5}>
      <Command>
        <Command.Input autoFocus placeholder="Join or manage teams..." />
        <Divider className="my-2" />
        <Command.Empty className="py-2">
          <Text color="muted">No teams found.</Text>
        </Command.Empty>
        <Command.Group>
          {teams.map((team) => (
            <Command.Item
              className="justify-between py-1 pr-1"
              disabled={teams.length === 1}
              key={team.id}
              onSelect={() => {
                setTeam(team.id, "leave");
                setOpen(false);
              }}
              value={team.name}
            >
              <Flex align="center" gap={2}>
                <Box
                  className="size-3 rounded"
                  style={{ backgroundColor: team.color }}
                />
                {team.name}
              </Flex>
              <Button
                className="px-2 border-border/80"
                color="tertiary"
                disabled={teams.length === 1}
                size="xs"
              >
                Leave
              </Button>
            </Command.Item>
          ))}

          {publicTeams.length > 0 && teams.length > 0 && (
            <Divider className="my-1.5" />
          )}

          {publicTeams.map((team) => (
            <Command.Item
              className="justify-between py-1 pr-1"
              key={team.id}
              onSelect={() => {
                setTeam(team.id, "join");
                setOpen(false);
              }}
              value={team.name}
            >
              <Flex align="center" gap={2}>
                <Box
                  className="size-3 rounded"
                  style={{ backgroundColor: team.color }}
                />
                {team.name}
              </Flex>
              <Button
                className="px-3 border-border/80"
                color="tertiary"
                size="xs"
              >
                Join team
              </Button>
            </Command.Item>
          ))}
        </Command.Group>
        {!hideManageTeams && (
          <>
            <Divider className="my-2" />
            <Command.Group>
              <Command.Item
                onSelect={() => {
                  router.push("/settings/workspace/teams");
                  setOpen(false);
                }}
              >
                <Flex align="center" gap={2}>
                  <TeamIcon className="h-4" />
                  <Text>Manage Teams</Text>
                </Flex>
              </Command.Item>
              <Command.Item
                onSelect={() => {
                  router.push("/settings/workspace/teams/create");
                  setOpen(false);
                }}
              >
                <Flex align="center" gap={2}>
                  <PlusIcon className="h-4" strokeWidth={3} />
                  <Text>Create new team</Text>
                </Flex>
              </Command.Item>
            </Command.Group>
          </>
        )}
      </Command>
    </Popover.Content>
  );
};

const Trigger = ({ children }: { children: ReactNode }) => (
  <Popover.Trigger asChild>{children}</Popover.Trigger>
);

TeamsMenu.Trigger = Trigger;
TeamsMenu.Items = Items;
