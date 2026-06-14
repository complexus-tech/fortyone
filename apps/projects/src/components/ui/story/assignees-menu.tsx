"use client";
import { AiIcon, CheckIcon } from "icons";
import {
  createContext,
  use,
  useDeferredValue,
  useState,
  type ReactNode,
  type UIEvent,
} from "react";
import { Avatar, Command, Flex, Popover, Text, Divider } from "ui";
import { useSession } from "@/lib/auth/client";
import {
  MEMBER_MENU_PAGE_SIZE,
  useMayaAssignee,
  useMembersInfinite,
} from "@/lib/hooks/members";
import { useTeamMembersInfinite } from "@/lib/hooks/team-members";
import { MenuLoadingSkeleton } from "../menu-loading-skeleton";

const AssigneesContext = createContext<{
  open: boolean;
  setOpen: (open: boolean) => void;
}>({
  open: false,
  setOpen: () => {},
});

const EMPTY_EXCLUDED_USERS: string[] = [];

export const useAssigneesMenu = () => {
  const { open, setOpen } = use(AssigneesContext);
  return { open, setOpen };
};

const Menu = ({ children }: { children: ReactNode }) => {
  const { open, setOpen } = useAssigneesMenu();
  return (
    <Popover onOpenChange={setOpen} open={open}>
      {children}
    </Popover>
  );
};

export const AssigneesMenu = ({ children }: { children: ReactNode }) => {
  const [open, setOpen] = useState(false);
  return (
    <AssigneesContext.Provider value={{ open, setOpen }}>
      <Menu>{children}</Menu>
    </AssigneesContext.Provider>
  );
};

const Trigger = ({ children }: { children: ReactNode }) => (
  <Popover.Trigger asChild>{children}</Popover.Trigger>
);

const Items = ({
  placeholder = "Assign to...",
  align,
  assigneeId,
  onAssigneeSelected,
  disallowEmptySelection = false,
  excludeUsers = EMPTY_EXCLUDED_USERS,
  teamId,
}: {
  placeholder?: string;
  align?: "start" | "end" | "center";
  disallowEmptySelection?: boolean;
  excludeUsers?: string[];
  assigneeId?: string | null;
  teamId?: string;
  onAssigneeSelected: (assigneeId: string | null) => void;
}) => {
  const { data: session } = useSession();
  const { open, setOpen } = useAssigneesMenu();
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const workspaceMembersQuery = useMembersInfinite(
    deferredQuery,
    MEMBER_MENU_PAGE_SIZE,
    open && !teamId,
  );
  const teamMembersQuery = useTeamMembersInfinite(
    teamId,
    deferredQuery,
    MEMBER_MENU_PAGE_SIZE,
    open && Boolean(teamId),
  );
  const membersQuery = teamId ? teamMembersQuery : workspaceMembersQuery;
  const mayaQuery = useMayaAssignee(open);
  const members =
    membersQuery.data?.pages.flatMap((page) => page.members) ?? [];
  const isLoadingMembers =
    membersQuery.isFetching && !membersQuery.isFetchingNextPage;
  const mayaAssignee = mayaQuery.data;
  const normalizedQuery = deferredQuery.trim().toLowerCase();
  const showMayaAssignee =
    mayaAssignee !== undefined &&
    !excludeUsers.includes(mayaAssignee.id) &&
    (normalizedQuery === "" ||
      "maya ai assistant".includes(normalizedQuery));
  const currentUserId = session?.user.id ?? null;
  const self = members.find(({ id }) => id === currentUserId);
  const visibleMembers = members.filter(
    ({ id }) =>
      !excludeUsers.includes(id) &&
      id !== currentUserId &&
      id !== mayaAssignee?.id,
  );
  const indexOffset =
    (disallowEmptySelection ? 0 : 1) +
    (showMayaAssignee ? 1 : 0) +
    (self ? 1 : 0);

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

  return (
    <Popover.Content align={align} className="w-80">
      <Command>
        <Command.Input
          autoFocus
          onValueChange={setQuery}
          placeholder={placeholder}
          value={query}
        />
        <Divider className="my-2" />
        {!isLoadingMembers ? (
          <Command.Empty className="py-2">
            <Text color="muted">No user found.</Text>
          </Command.Empty>
        ) : null}
        <Command.Group
          className="max-h-80 overflow-y-auto md:max-h-100"
          onScroll={handleScroll}
        >
          {isLoadingMembers ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton avatar rows={5} />
            </Command.Loading>
          ) : null}

          {!isLoadingMembers && (
            <>
              {!disallowEmptySelection ? (
                <Command.Item
                  active={!assigneeId}
                  className="justify-between opacity-70"
                  onSelect={() => {
                    if (assigneeId) {
                      onAssigneeSelected(null);
                    }
                    setOpen(false);
                  }}
                >
                  <Flex align="center" gap={2}>
                    <Avatar
                      className="text-foreground/80"
                      color="primary"
                      size="sm"
                    />
                    <Text className="max-w-40 truncate">Unassigned</Text>
                  </Flex>
                  <Flex align="center" gap={1}>
                    {!assigneeId && (
                      <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                    )}
                    <Text color="muted">0</Text>
                  </Flex>
                </Command.Item>
              ) : null}
              {showMayaAssignee && mayaAssignee ? (
                <Command.Item
                  active={mayaAssignee.id === assigneeId}
                  className="justify-between"
                  onSelect={() => {
                    if (mayaAssignee.id !== assigneeId) {
                      onAssigneeSelected(mayaAssignee.id);
                    }
                    setOpen(false);
                  }}
                  value="Maya AI assistant"
                >
                  <Flex align="center" gap={2}>
                    <Flex
                      align="center"
                      className="size-6 rounded-full bg-black text-white dark:bg-white dark:text-black"
                      justify="center"
                    >
                      <AiIcon className="size-3.5 text-current" />
                    </Flex>
                    <Text className="max-w-48 truncate">
                      Maya{" "}
                      <Text as="span" color="muted">
                        (AI)
                      </Text>
                    </Text>
                  </Flex>
                  <Flex align="center" gap={1}>
                    {mayaAssignee.id === assigneeId && (
                      <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                    )}
                    <Text color="muted">
                      {disallowEmptySelection ? 0 : 1}
                    </Text>
                  </Flex>
                </Command.Item>
              ) : null}
              {self ? (
                <Command.Item
                  active={self.id === assigneeId}
                  className="justify-between"
                  onSelect={() => {
                    if (self.id !== assigneeId) {
                      onAssigneeSelected(self.id);
                    }
                    setOpen(false);
                  }}
                  value={self.fullName || self.username || self.email}
                >
                  <Flex align="center" gap={2}>
                    <Avatar
                      color="primary"
                      name={self.fullName}
                      size="sm"
                      src={self.avatarUrl}
                    />
                    <Text className="max-w-48 truncate">
                      {self.fullName || self.username}{" "}
                      <Text as="span" color="muted">
                        (You)
                      </Text>
                    </Text>
                  </Flex>
                  <Flex align="center" gap={1}>
                    {self.id === assigneeId && (
                      <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                    )}
                    <Text color="muted">{disallowEmptySelection ? 0 : 1}</Text>
                  </Flex>
                </Command.Item>
              ) : null}
            </>
          )}
          {visibleMembers.map(({ id, fullName, username, avatarUrl }, idx) => (
            <Command.Item
              active={id === assigneeId}
              className="justify-between"
              key={id}
              onSelect={() => {
                if (id !== assigneeId) {
                  onAssigneeSelected(id);
                }
                setOpen(false);
              }}
              value={fullName || username}
            >
              <Flex align="center" gap={2}>
                <Avatar
                  color="primary"
                  name={fullName || username}
                  size="sm"
                  src={avatarUrl}
                />
                <Text className="max-w-48 truncate">
                  {fullName || username}
                </Text>
              </Flex>
              <Flex align="center" gap={1}>
                {id === assigneeId && (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                )}
                <Text color="muted">{idx + indexOffset}</Text>
              </Flex>
            </Command.Item>
          ))}
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

AssigneesMenu.Trigger = Trigger;
AssigneesMenu.Items = Items;
