"use client";

import { CheckIcon } from "icons";
import {
  createContext,
  use,
  useDeferredValue,
  useState,
  type ReactNode,
  type UIEvent,
} from "react";
import { Avatar, Command, Divider, Flex, Popover, Text } from "ui";
import { MEMBER_MENU_PAGE_SIZE } from "@/lib/hooks/members";
import { useTeamMembersInfinite } from "@/lib/hooks/team-members";
import { MenuLoadingSkeleton } from "../menu-loading-skeleton";

const CollaboratorsContext = createContext<{
  open: boolean;
  setOpen: (open: boolean) => void;
}>({
  open: false,
  setOpen: () => {},
});

const useCollaboratorsMenu = () => use(CollaboratorsContext);

const Menu = ({ children }: { children: ReactNode }) => {
  const { open, setOpen } = useCollaboratorsMenu();
  return (
    <Popover onOpenChange={setOpen} open={open}>
      {children}
    </Popover>
  );
};

export const CollaboratorsMenu = ({ children }: { children: ReactNode }) => {
  const [open, setOpen] = useState(false);

  return (
    <CollaboratorsContext.Provider value={{ open, setOpen }}>
      <Menu>{children}</Menu>
    </CollaboratorsContext.Provider>
  );
};

const Trigger = ({ children }: { children: ReactNode }) => (
  <Popover.Trigger asChild>{children}</Popover.Trigger>
);

const Items = ({
  align,
  assigneeId,
  collaboratorIds,
  onCollaboratorsChange,
  teamId,
}: {
  align?: "start" | "end" | "center";
  assigneeId?: string | null;
  collaboratorIds: string[];
  onCollaboratorsChange: (collaboratorIds: string[]) => void;
  teamId: string;
}) => {
  const { open } = useCollaboratorsMenu();
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const membersQuery = useTeamMembersInfinite(
    teamId,
    deferredQuery,
    MEMBER_MENU_PAGE_SIZE,
    open,
  );
  const members =
    membersQuery.data?.pages.flatMap((page) => page.members) ?? [];
  const selectedIds = new Set(collaboratorIds);
  const visibleMembers = members.filter(({ id }) => id !== assigneeId);
  const isLoading = membersQuery.isFetching && !membersQuery.isFetchingNextPage;

  const handleScroll = (event: UIEvent<HTMLDivElement>) => {
    const target = event.currentTarget;
    const distanceToBottom =
      target.scrollHeight - target.scrollTop - target.clientHeight;
    if (
      distanceToBottom <= 80 &&
      membersQuery.hasNextPage &&
      !membersQuery.isFetchingNextPage
    ) {
      void membersQuery.fetchNextPage();
    }
  };

  const toggleCollaborator = (memberId: string) => {
    onCollaboratorsChange(
      selectedIds.has(memberId)
        ? collaboratorIds.filter((id) => id !== memberId)
        : [...collaboratorIds, memberId],
    );
  };

  return (
    <Popover.Content align={align} className="w-80">
      <Command>
        <Command.Input
          autoFocus
          onValueChange={setQuery}
          placeholder="Search collaborators..."
          value={query}
        />
        <Divider className="my-2" />
        {!isLoading ? (
          <Command.Empty className="py-2">
            <Text color="muted">No user found.</Text>
          </Command.Empty>
        ) : null}
        <Command.Group
          className="max-h-80 overflow-y-auto md:max-h-100"
          onScroll={handleScroll}
        >
          {isLoading ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton avatar rows={5} />
            </Command.Loading>
          ) : null}
          {visibleMembers.map(({ id, fullName, username, avatarUrl }) => {
            const selected = selectedIds.has(id);
            const name = fullName || username;
            return (
              <Command.Item
                active={selected}
                className="justify-between"
                key={id}
                onSelect={() => {
                  toggleCollaborator(id);
                }}
                value={name}
              >
                <Flex align="center" gap={2}>
                  <Avatar
                    color="primary"
                    name={name}
                    size="sm"
                    src={avatarUrl}
                  />
                  <Text className="max-w-48 truncate">{name}</Text>
                </Flex>
                {selected ? (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                ) : null}
              </Command.Item>
            );
          })}
          {membersQuery.isFetchingNextPage ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton avatar rows={2} />
            </Command.Loading>
          ) : null}
        </Command.Group>
      </Command>
    </Popover.Content>
  );
};

CollaboratorsMenu.Trigger = Trigger;
CollaboratorsMenu.Items = Items;
