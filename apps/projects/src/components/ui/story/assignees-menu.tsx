"use client";
import { CheckIcon, LoadingIcon } from "icons";
import { createContext, useContext, useState, type ReactNode } from "react";
import { Avatar, Command, Flex, Popover, Text, Divider } from "ui";
import { useSession } from "next-auth/react";
import { useMembers } from "@/lib/hooks/members";
import { useTeamMembers } from "@/lib/hooks/team-members";

const AssigneesContext = createContext<{
  open: boolean;
  setOpen: (open: boolean) => void;
}>({
  open: false,
  setOpen: () => {},
});

export const useAssigneesMenu = () => {
  const { open, setOpen } = useContext(AssigneesContext);
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
  excludeUsers = [],
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
  const { setOpen } = useAssigneesMenu();
  const { data: allMembers = [], isPending: isLoadingMembers } = useMembers();
  const { data: teamMembers = [], isPending: isLoadingTeamMembers } =
    useTeamMembers(teamId);

  const members = teamId ? teamMembers : allMembers;
  const isPending = teamId ? isLoadingTeamMembers : isLoadingMembers;
  const self = members.find(({ id }) => id === session?.user?.id);

  return (
    <Popover.Content align={align} className="w-80">
      <Command>
        <Command.Input autoFocus placeholder={placeholder} />
        <Divider className="my-2" />
        <Command.Empty className="py-2">
          <Text color="muted">No user found.</Text>
        </Command.Empty>
        <Command.Group>
          {isPending ? (
            <Command.Loading className="p-2">
              <Text className="flex items-center gap-2" color="muted">
                <LoadingIcon className="animate-spin" />
                Please wait...
              </Text>
            </Command.Loading>
          ) : null}

          {!isPending && (
            <>
              {!disallowEmptySelection ? (
                <>
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
                        className="text-dark/80 dark:text-gray-200"
                        color="primary"
                        size="sm"
                      />
                      <Text className="max-w-[10rem] truncate">Unassigned</Text>
                    </Flex>
                    <Flex align="center" gap={1}>
                      {!assigneeId && (
                        <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                      )}
                      <Text color="muted">0</Text>
                    </Flex>
                  </Command.Item>
                  {members.length > 0 && <Divider className="my-2" />}
                </>
              ) : null}
              <Command.Item
                active={self?.id === assigneeId}
                className="justify-between"
                onSelect={() => {
                  if (self?.id !== assigneeId) {
                    onAssigneeSelected(self?.id ?? null);
                  }
                  setOpen(false);
                }}
              >
                <Flex align="center" gap={2}>
                  <Avatar
                    color="primary"
                    name={self?.fullName}
                    size="sm"
                    src={self?.avatarUrl}
                  />
                  <Text className="max-w-[12rem] truncate">
                    {self?.fullName || self?.username}{" "}
                    <Text as="span" color="muted">
                      (You)
                    </Text>
                  </Text>
                </Flex>
                <Flex align="center" gap={1}>
                  {self?.id === assigneeId && (
                    <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                  )}
                  <Text color="muted">{1}</Text>
                </Flex>
              </Command.Item>
            </>
          )}
          {members
            .filter(
              ({ id }) =>
                !excludeUsers.includes(id) && id !== session?.user?.id,
            )
            .map(({ id, fullName, username, avatarUrl }, idx) => (
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
              >
                <Flex align="center" gap={2}>
                  <Avatar
                    color="primary"
                    name={fullName || username}
                    size="sm"
                    src={avatarUrl}
                  />
                  <Text className="max-w-[12rem] truncate">
                    {fullName || username}
                  </Text>
                </Flex>
                <Flex align="center" gap={1}>
                  {id === assigneeId && (
                    <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                  )}
                  <Text color="muted">{idx + 2}</Text>
                </Flex>
              </Command.Item>
            ))}
        </Command.Group>
      </Command>
    </Popover.Content>
  );
};

AssigneesMenu.Trigger = Trigger;
AssigneesMenu.Items = Items;
