"use client";
import { CheckIcon } from "icons";
import { createContext, useContext, useState, type ReactNode } from "react";
import { Avatar, Command, Flex, Popover, Text, Divider } from "ui";
import { useMembers } from "@/lib/hooks/members";

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
}: {
  placeholder?: string;
  align?: "start" | "end" | "center";
  disallowEmptySelection?: boolean;
  excludeUsers?: string[];
  assigneeId?: string | null;
  onAssigneeSelected: (assigneeId: string | null) => void;
}) => {
  const { setOpen } = useAssigneesMenu();
  const { data: members = [] } = useMembers();

  return (
    <Popover.Content align={align} className="w-72">
      <Command>
        <Command.Input autoFocus placeholder={placeholder} />
        <Divider className="my-2" />
        <Command.Empty className="py-2">
          <Text color="muted">No user found.</Text>
        </Command.Empty>
        <Command.Group>
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
                  <Text className="max-w-[10rem] truncate">No assignee</Text>
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

          {members
            .filter(({ id }) => !excludeUsers.includes(id))
            .map(({ id, fullName, avatarUrl }, idx) => (
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
                    name={fullName}
                    size="sm"
                    src={avatarUrl}
                  />
                  <Text className="max-w-[10rem] truncate">{fullName}</Text>
                </Flex>
                <Flex align="center" gap={1}>
                  {id === assigneeId && (
                    <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                  )}
                  <Text color="muted">{idx + 1}</Text>
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
