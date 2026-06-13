"use client";

import { Command, Flex, Popover, Text, Divider, Button, Box } from "ui";
import {
  createContext,
  use,
  useDeferredValue,
  useState,
  type ReactNode,
  type UIEvent,
} from "react";
import { PlusIcon, TeamIcon } from "icons";
import { useRouter } from "next/navigation";
import {
  TEAM_MENU_PAGE_SIZE,
  usePublicTeamsInfinite,
  useTeamsInfinite,
} from "@/modules/teams/hooks/teams";
import { useWorkspacePath } from "@/hooks";
import { MenuLoadingSkeleton } from "./menu-loading-skeleton";

type TeamContextType = {
  open: boolean;
  setOpen: (open: boolean) => void;
};

const TeamsContext = createContext<TeamContextType>({
  open: false,
  setOpen: () => {},
});

export const useTeamsMenu = () => {
  const context = use(TeamsContext);
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
  const { withWorkspace } = useWorkspacePath();
  const { open, setOpen } = useTeamsMenu();
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const joinedTeamsQuery = useTeamsInfinite(
    deferredQuery,
    TEAM_MENU_PAGE_SIZE,
    open,
  );
  const publicTeamsQuery = usePublicTeamsInfinite(
    deferredQuery,
    TEAM_MENU_PAGE_SIZE,
    open,
  );
  const teams =
    joinedTeamsQuery.data?.pages.flatMap((page) => page.teams) ?? [];
  const publicTeams =
    publicTeamsQuery.data?.pages.flatMap((page) => page.teams) ?? [];
  const canLeaveTeam =
    teams.length > 1 || Boolean(joinedTeamsQuery.hasNextPage);
  const isInitialLoading =
    joinedTeamsQuery.isPending || publicTeamsQuery.isPending;

  const handleScroll = (event: UIEvent<HTMLDivElement>) => {
    const target = event.currentTarget;
    const distanceToBottom =
      target.scrollHeight - target.scrollTop - target.clientHeight;

    if (distanceToBottom > 80) {
      return;
    }

    if (joinedTeamsQuery.hasNextPage && !joinedTeamsQuery.isFetchingNextPage) {
      void joinedTeamsQuery.fetchNextPage();
    }
    if (publicTeamsQuery.hasNextPage && !publicTeamsQuery.isFetchingNextPage) {
      void publicTeamsQuery.fetchNextPage();
    }
  };

  return (
    <Popover.Content align="start" className="w-72" sideOffset={5}>
      <Command>
        <Command.Input
          onValueChange={setQuery}
          placeholder="Join or manage teams..."
          value={query}
        />
        <Divider className="my-2" />
        {!isInitialLoading ? (
          <Command.Empty className="py-2">
            <Text color="muted">No teams found.</Text>
          </Command.Empty>
        ) : null}
        <Command.Group
          className="max-h-80 overflow-y-auto"
          onScroll={handleScroll}
        >
          {joinedTeamsQuery.isPending ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton rows={4} />
            </Command.Loading>
          ) : null}
          {teams.map((team) => (
            <Command.Item
              className="justify-between py-1 pr-1"
              disabled={!canLeaveTeam}
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
                className="border-border/80 px-2"
                color="tertiary"
                disabled={!canLeaveTeam}
                size="xs"
              >
                Leave
              </Button>
            </Command.Item>
          ))}
          {joinedTeamsQuery.isFetchingNextPage ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton rows={2} />
            </Command.Loading>
          ) : null}

          {publicTeams.length > 0 && teams.length > 0 && (
            <Divider className="my-1.5" />
          )}

          {publicTeamsQuery.isPending ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton rows={4} />
            </Command.Loading>
          ) : null}

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
                className="border-border/80 px-3"
                color="tertiary"
                size="xs"
              >
                Join team
              </Button>
            </Command.Item>
          ))}
          {publicTeamsQuery.isFetchingNextPage ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton rows={2} />
            </Command.Loading>
          ) : null}
        </Command.Group>
        {!hideManageTeams && (
          <>
            <Divider className="my-2" />
            <Command.Group>
              <Command.Item
                onSelect={() => {
                  router.push(withWorkspace("/settings/workspace/teams"));
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
                  router.push(
                    withWorkspace("/settings/workspace/teams/create"),
                  );
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
