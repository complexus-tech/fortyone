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
import {
  Avatar,
  Command,
  Divider,
  Flex,
  Menu as DropdownMenu,
  Popover,
  Text,
} from "ui";
import { useSession } from "@/lib/auth/client";
import { MEMBER_MENU_PAGE_SIZE, useMembersInfinite } from "@/lib/hooks/members";
import { useTeamMembersInfinite } from "@/lib/hooks/team-members";
import { MenuLoadingSkeleton } from "../menu-loading-skeleton";
import {
  isPropertySelectionActive,
  shouldApplyPropertySelection,
  type PropertyMenuSelectionMode,
} from "./property-menu-selection";

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

const PopoverMenu = ({ children }: { children: ReactNode }) => {
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
      <PopoverMenu>{children}</PopoverMenu>
    </AssigneesContext.Provider>
  );
};

const Trigger = ({ children }: { children: ReactNode }) => (
  <Popover.Trigger asChild>{children}</Popover.Trigger>
);

type AssigneePickerProps = {
  assigneeId?: string | null;
  disallowEmptySelection?: boolean;
  excludeUsers?: string[];
  onAssigneeSelected: (assigneeId: string | null) => void;
  placeholder?: string;
  selectionMode?: PropertyMenuSelectionMode;
  teamId?: string;
};

const AssigneePickerContent = ({
  placeholder = "Assign to...",
  assigneeId,
  onAssigneeSelected,
  disallowEmptySelection = false,
  excludeUsers = EMPTY_EXCLUDED_USERS,
  open,
  selectionMode = "single",
  setOpen,
  teamId,
}: AssigneePickerProps & {
  open: boolean;
  setOpen: (open: boolean) => void;
}) => {
  const { data: session } = useSession();
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
  const members =
    membersQuery.data?.pages.flatMap((page) => page.members) ?? [];
  const excludedUserIds = new Set(excludeUsers);
  const isLoadingMembers =
    membersQuery.isFetching && !membersQuery.isFetchingNextPage;
  const currentUserId = session?.user.id ?? null;
  const self = members.find(({ id }) => id === currentUserId);
  const visibleMembers = members.filter(
    ({ id }) => !excludedUserIds.has(id) && id !== currentUserId,
  );
  const indexOffset = (disallowEmptySelection ? 0 : 1) + (self ? 1 : 0);

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
                active={isPropertySelectionActive(
                  selectionMode,
                  assigneeId ?? null,
                  null,
                )}
                className="justify-between opacity-70"
                onSelect={() => {
                  if (
                    shouldApplyPropertySelection(
                      selectionMode,
                      assigneeId ?? null,
                      null,
                    )
                  ) {
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
                  {isPropertySelectionActive(
                    selectionMode,
                    assigneeId ?? null,
                    null,
                  ) && <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />}
                  <Text color="muted">0</Text>
                </Flex>
              </Command.Item>
            ) : null}
            {self ? (
              <Command.Item
                active={isPropertySelectionActive(
                  selectionMode,
                  assigneeId ?? null,
                  self.id,
                )}
                className="justify-between"
                onSelect={() => {
                  if (
                    shouldApplyPropertySelection(
                      selectionMode,
                      assigneeId ?? null,
                      self.id,
                    )
                  ) {
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
                  {isPropertySelectionActive(
                    selectionMode,
                    assigneeId ?? null,
                    self.id,
                  ) && <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />}
                  <Text color="muted">{disallowEmptySelection ? 0 : 1}</Text>
                </Flex>
              </Command.Item>
            ) : null}
          </>
        )}
        {visibleMembers.map(({ id, fullName, username, avatarUrl }, idx) => (
          <Command.Item
            active={isPropertySelectionActive(
              selectionMode,
              assigneeId ?? null,
              id,
            )}
            className="justify-between"
            key={id}
            onSelect={() => {
              if (
                shouldApplyPropertySelection(
                  selectionMode,
                  assigneeId ?? null,
                  id,
                )
              ) {
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
              <Text className="max-w-48 truncate">{fullName || username}</Text>
            </Flex>
            <Flex align="center" gap={1}>
              {isPropertySelectionActive(
                selectionMode,
                assigneeId ?? null,
                id,
              ) && <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />}
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
  );
};

const Items = ({
  align,
  ...pickerProps
}: AssigneePickerProps & {
  align?: "start" | "end" | "center";
}) => {
  const { open, setOpen } = useAssigneesMenu();

  return (
    <Popover.Content align={align} className="w-80">
      <AssigneePickerContent {...pickerProps} open={open} setOpen={setOpen} />
    </Popover.Content>
  );
};

const SubMenu = ({
  children,
  ...pickerProps
}: AssigneePickerProps & { children: ReactNode }) => {
  const [open, setOpen] = useState(false);

  return (
    <DropdownMenu.SubMenu onOpenChange={setOpen} open={open}>
      <DropdownMenu.SubTrigger className="justify-between gap-4">
        {children}
      </DropdownMenu.SubTrigger>
      <DropdownMenu.SubItems className="w-80 overflow-hidden">
        <AssigneePickerContent {...pickerProps} open={open} setOpen={setOpen} />
      </DropdownMenu.SubItems>
    </DropdownMenu.SubMenu>
  );
};

AssigneesMenu.Trigger = Trigger;
AssigneesMenu.Items = Items;
AssigneesMenu.SubMenu = SubMenu;
