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
import { ChevronRightIcon, PlusIcon, TeamIcon } from "icons";
import { useRouter } from "next/navigation";
import {
  TEAM_MENU_PAGE_SIZE,
  useJoinedTeamsInfinite,
  usePublicTeamsInfinite,
} from "@/modules/teams/hooks/teams";
import { useWorkspacePath } from "@/hooks";
import { MenuLoadingSkeleton } from "./menu-loading-skeleton";

const INITIAL_TEAM_MENU_SKELETON_ROWS = 2;

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
  readOnly = false,
  setTeam,
}: {
  hideManageTeams?: boolean;
  readOnly?: boolean;
  setTeam: (teamId: string, action: "join" | "leave") => void;
}) => {
  const router = useRouter();
  const { withWorkspace } = useWorkspacePath();
  const { open, setOpen } = useTeamsMenu();
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const joinedTeamsQuery = useJoinedTeamsInfinite(
    deferredQuery,
    TEAM_MENU_PAGE_SIZE,
    open,
  );
  const publicTeamsQuery = usePublicTeamsInfinite(
    deferredQuery,
    TEAM_MENU_PAGE_SIZE,
    open && !readOnly,
  );
  const joinedTeams =
    joinedTeamsQuery.data?.pages.flatMap((page) => page.teams) ?? [];
  const publicTeams =
    publicTeamsQuery.data?.pages.flatMap((page) => page.teams) ?? [];
  const isInitialLoading =
    (joinedTeamsQuery.isFetching && !joinedTeamsQuery.isFetchingNextPage) ||
    (!readOnly &&
      publicTeamsQuery.isFetching &&
      !publicTeamsQuery.isFetchingNextPage);

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
        <Command.List
          className="mt-0 max-h-80 w-full overflow-y-auto border-0 bg-transparent py-0 shadow-none backdrop-blur-none dark:bg-transparent"
          onScroll={handleScroll}
        >
          {!hideManageTeams ? (
            <>
              <Command.Group>
                <Command.Item
                  onSelect={() => {
                    router.push(withWorkspace("/settings/workspace/teams"));
                    setOpen(false);
                  }}
                >
                  <Flex align="center" gap={2}>
                    <TeamIcon className="h-5 w-auto" />
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
                    <PlusIcon className="h-4" strokeWidth={2} />
                    <Text>Create new team</Text>
                  </Flex>
                </Command.Item>
              </Command.Group>
              <Divider className="my-2" />
            </>
          ) : null}
          {isInitialLoading ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton rows={INITIAL_TEAM_MENU_SKELETON_ROWS} />
            </Command.Loading>
          ) : null}
          {!isInitialLoading &&
          joinedTeams.length === 0 &&
          (readOnly || publicTeams.length === 0) ? (
            <Box className="px-3 py-2">
              <Text color="muted">No teams found.</Text>
            </Box>
          ) : null}
          {joinedTeams.length > 0 ? (
            <Command.Group
              heading={
                <Text className="mb-1 px-2" color="muted" fontWeight="medium">
                  Your teams
                </Text>
              }
            >
              {joinedTeams.map((team) => (
                <Command.Item
                  className="justify-between py-1 pr-1"
                  key={team.id}
                  onSelect={() => {
                    router.push(withWorkspace(`/teams/${team.id}/stories`));
                    setOpen(false);
                  }}
                  value={`${team.name} ${team.id}`}
                >
                  <Flex align="center" className="min-w-0" gap={2}>
                    <Box
                      className="size-3 shrink-0 rounded"
                      style={{ backgroundColor: team.color }}
                    />
                    <span className="truncate">{team.name}</span>
                  </Flex>
                  <ChevronRightIcon className="h-4 w-auto shrink-0" />
                </Command.Item>
              ))}
            </Command.Group>
          ) : null}
          {joinedTeamsQuery.isFetchingNextPage ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton rows={2} />
            </Command.Loading>
          ) : null}

          {!readOnly && publicTeams.length > 0 && joinedTeams.length > 0 ? (
            <Divider className="my-1.5" />
          ) : null}

          {!readOnly && publicTeams.length > 0 ? (
            <Command.Group
              heading={
                <Text className="mb-1 px-2" color="muted" fontWeight="medium">
                  Join a team
                </Text>
              }
            >
              {publicTeams.map((team) => (
                <Command.Item
                  className="justify-between py-1 pr-1"
                  key={team.id}
                  onSelect={() => {
                    setTeam(team.id, "join");
                    setOpen(false);
                  }}
                  value={`${team.name} ${team.id}`}
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
            </Command.Group>
          ) : null}
          {!readOnly && publicTeamsQuery.isFetchingNextPage ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton rows={2} />
            </Command.Loading>
          ) : null}
        </Command.List>
      </Command>
    </Popover.Content>
  );
};

const Trigger = ({ children }: { children: ReactNode }) => (
  <Popover.Trigger asChild>{children}</Popover.Trigger>
);

TeamsMenu.Trigger = Trigger;
TeamsMenu.Items = Items;
