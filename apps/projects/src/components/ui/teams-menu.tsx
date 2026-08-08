"use client";

import type { ReactNode, UIEvent } from "react";
import { createContext, use, useDeferredValue, useState } from "react";
import { useRouter } from "next/navigation";
import { ChevronRightIcon, PinIcon, PlusIcon, TeamIcon } from "icons";
import { Box, Divider, Flex, Menu as DropdownMenu, Text, Tooltip } from "ui";
import { useWorkspacePath } from "@/hooks";
import {
  TEAM_MENU_PAGE_SIZE,
  usePublicTeamsInfinite,
} from "@/modules/teams/hooks/teams";
import type { Team } from "@/modules/teams/types";
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

const MenuRoot = ({ children }: { children: ReactNode }) => {
  const { open, setOpen } = useTeamsMenu();

  return (
    <DropdownMenu onOpenChange={setOpen} open={open}>
      {children}
    </DropdownMenu>
  );
};

export const TeamsMenu = ({ children }: { children: ReactNode }) => {
  const [open, setOpen] = useState(false);

  return (
    <TeamsContext.Provider value={{ open, setOpen }}>
      <MenuRoot>{children}</MenuRoot>
    </TeamsContext.Provider>
  );
};

const YourTeamsSubMenu = ({
  onPinTeam,
  overflowTeams,
  setTeam,
}: {
  onPinTeam: (teamId: string) => void;
  overflowTeams: Team[];
  setTeam: (teamId: string, action: "join" | "leave") => void;
}) => {
  const router = useRouter();
  const { withWorkspace } = useWorkspacePath();
  const { setOpen } = useTeamsMenu();
  const [isOpen, setIsOpen] = useState(false);
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query.trim().toLocaleLowerCase());
  const filteredTeams = overflowTeams.filter((team) =>
    team.name.toLocaleLowerCase().includes(deferredQuery),
  );

  return (
    <DropdownMenu.SubMenu
      onOpenChange={(open) => {
        setIsOpen(open);
        if (!open) setQuery("");
      }}
      open={isOpen}
    >
      <DropdownMenu.SubTrigger className="justify-between gap-4">
        <Flex align="center" className="min-w-0" gap={2}>
          <TeamIcon className="h-5 w-auto shrink-0" />
          <Text>Your Teams</Text>
        </Flex>
        <ChevronRightIcon
          className="text-text-muted h-4 w-auto shrink-0"
          strokeWidth={2.4}
        />
      </DropdownMenu.SubTrigger>
      <DropdownMenu.SubItems className="w-80 max-w-[calc(100vw-2rem)]">
        <Box className="px-3 pt-0.5 pb-1.5">
          <DropdownMenu.Input
            aria-label="Search your teams"
            autoFocus
            onChange={(event) => {
              setQuery(event.target.value);
            }}
            placeholder="Find your teams..."
            value={query}
          />
        </Box>
        <Divider className="my-1.5" />
        <DropdownMenu.Group className="max-h-72 overflow-y-auto">
          {filteredTeams.length === 0 ? (
            <Text className="px-2 py-2" color="muted">
              No teams found.
            </Text>
          ) : null}
          {filteredTeams.map((team) => (
            <Flex
              align="center"
              aria-label={`${team.name} actions`}
              className="gap-1 py-0.5"
              key={team.id}
              role="group"
            >
              <DropdownMenu.Item
                aria-label={`Open ${team.name}`}
                className="min-w-0 flex-1"
                onSelect={() => {
                  router.push(withWorkspace(`/teams/${team.id}/stories`));
                  setOpen(false);
                }}
              >
                <Flex align="center" className="min-w-0" gap={2}>
                  <Box
                    className="size-3 shrink-0 rounded"
                    style={{ backgroundColor: team.color }}
                  />
                  <Text className="truncate">{team.name}</Text>
                </Flex>
              </DropdownMenu.Item>
              <DropdownMenu.Item
                aria-label={`Leave ${team.name}`}
                className="border-border/80 h-7.5 w-auto shrink-0 rounded-xl border px-2 py-0 text-[0.95rem]"
                onSelect={() => {
                  setTeam(team.id, "leave");
                  setOpen(false);
                }}
              >
                Leave
              </DropdownMenu.Item>
              <Tooltip title="Pin">
                <DropdownMenu.Item
                  aria-label={`Pin ${team.name}`}
                  className="border-border/80 h-7.5 w-7.5 shrink-0 justify-center rounded-xl border p-0"
                  onSelect={() => {
                    onPinTeam(team.id);
                    setOpen(false);
                  }}
                >
                  <PinIcon className="h-4 w-auto" strokeWidth={2} />
                </DropdownMenu.Item>
              </Tooltip>
            </Flex>
          ))}
        </DropdownMenu.Group>
      </DropdownMenu.SubItems>
    </DropdownMenu.SubMenu>
  );
};

const Items = ({
  hideManageTeams,
  onPinTeam,
  overflowTeams,
  readOnly = false,
  setTeam,
}: {
  hideManageTeams?: boolean;
  onPinTeam: (teamId: string) => void;
  overflowTeams: Team[];
  readOnly?: boolean;
  setTeam: (teamId: string, action: "join" | "leave") => void;
}) => {
  const router = useRouter();
  const { withWorkspace } = useWorkspacePath();
  const { open, setOpen } = useTeamsMenu();
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const publicTeamsQuery = usePublicTeamsInfinite(
    deferredQuery,
    TEAM_MENU_PAGE_SIZE,
    open && !readOnly,
  );
  const publicTeams =
    publicTeamsQuery.data?.pages.flatMap((page) => page.teams) ?? [];
  const isInitialLoading =
    publicTeamsQuery.isFetching && !publicTeamsQuery.isFetchingNextPage;
  const hasActionGroup = !hideManageTeams || overflowTeams.length > 0;

  const handleScroll = (event: UIEvent<HTMLDivElement>) => {
    const target = event.currentTarget;
    const distanceToBottom =
      target.scrollHeight - target.scrollTop - target.clientHeight;

    if (distanceToBottom > 80) return;

    if (publicTeamsQuery.hasNextPage && !publicTeamsQuery.isFetchingNextPage) {
      void publicTeamsQuery.fetchNextPage();
    }
  };

  return (
    <DropdownMenu.Items
      align="start"
      className="w-72"
      onCloseAutoFocus={() => {
        setQuery("");
      }}
      sideOffset={5}
    >
      {!readOnly ? (
        <>
          <Box className="px-3 pt-0.5 pb-1.5">
            <DropdownMenu.Input
              aria-label="Search teams"
              onChange={(event) => {
                setQuery(event.target.value);
              }}
              placeholder="Find a team..."
              value={query}
            />
          </Box>
          <Divider className="my-1.5" />
        </>
      ) : null}

      {hasActionGroup ? (
        <DropdownMenu.Group>
          {!hideManageTeams ? (
            <>
              <DropdownMenu.Item
                onSelect={() => {
                  router.push(withWorkspace("/settings/workspace/teams"));
                  setOpen(false);
                }}
              >
                <TeamIcon className="h-5 w-auto" />
                <Text>Manage Teams</Text>
              </DropdownMenu.Item>
              <DropdownMenu.Item
                onSelect={() => {
                  router.push(
                    withWorkspace("/settings/workspace/teams/create"),
                  );
                  setOpen(false);
                }}
              >
                <PlusIcon className="h-4" strokeWidth={2} />
                <Text>Create new team</Text>
              </DropdownMenu.Item>
            </>
          ) : null}
          {overflowTeams.length > 0 ? (
            <YourTeamsSubMenu
              onPinTeam={onPinTeam}
              overflowTeams={overflowTeams}
              setTeam={setTeam}
            />
          ) : null}
        </DropdownMenu.Group>
      ) : null}

      {!readOnly && hasActionGroup ? <DropdownMenu.Separator /> : null}

      {!readOnly ? (
        <Box
          className="max-h-80 w-full overflow-y-auto"
          onScroll={handleScroll}
        >
          {isInitialLoading ? (
            <Box className="p-2">
              <MenuLoadingSkeleton rows={INITIAL_TEAM_MENU_SKELETON_ROWS} />
            </Box>
          ) : null}
          {!isInitialLoading && publicTeams.length === 0 ? (
            <Box className="px-3 py-2">
              <Text color="muted">No teams found.</Text>
            </Box>
          ) : null}
          {publicTeams.length > 0 ? (
            <DropdownMenu.Group>
              {publicTeams.map((team) => (
                <DropdownMenu.Item
                  className="justify-between py-1 pr-1"
                  key={team.id}
                  onSelect={() => {
                    setTeam(team.id, "join");
                    setOpen(false);
                  }}
                >
                  <Flex align="center" className="min-w-0" gap={2}>
                    <Box
                      className="size-3 shrink-0 rounded"
                      style={{ backgroundColor: team.color }}
                    />
                    <Text className="truncate">{team.name}</Text>
                  </Flex>
                  <span className="border-border/80 bg-surface flex h-7.5 shrink-0 items-center rounded-xl border px-2 text-[0.95rem]">
                    Join team
                  </span>
                </DropdownMenu.Item>
              ))}
            </DropdownMenu.Group>
          ) : null}
          {publicTeamsQuery.isFetchingNextPage ? (
            <Box className="p-2">
              <MenuLoadingSkeleton rows={2} />
            </Box>
          ) : null}
        </Box>
      ) : null}
    </DropdownMenu.Items>
  );
};

const Trigger = ({ children }: { children: ReactNode }) => (
  <DropdownMenu.Button>{children}</DropdownMenu.Button>
);

TeamsMenu.Trigger = Trigger;
TeamsMenu.Items = Items;
