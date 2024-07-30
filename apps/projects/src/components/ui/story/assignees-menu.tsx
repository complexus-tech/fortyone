"use client";
import { CheckIcon } from "icons";
import { createContext, useContext, useState, type ReactNode } from "react";
import { Avatar, Command, Flex, Popover, Text, Divider } from "ui";

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
    <Popover open={open} onOpenChange={setOpen}>
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
  placeholder = "Assign user...",
  align,
  onAssigneeSelected,
}: {
  placeholder?: string;
  align?: "start" | "end" | "center";
  onAssigneeSelected: (assignee: string) => void;
}) => {
  const { setOpen } = useAssigneesMenu();
  const users = [
    {
      name: "Joseph Mukorivo",
      avatar:
        "https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo",
    },
    {
      name: "Jane Doe",
      avatar:
        "https://images.unsplash.com/photo-1677576874778-df95ea6ff733?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHx0b3BpYy1mZWVkfDI4fHRvd0paRnNrcEdnfHxlbnwwfHx8fHw%3D&auto=format&fit=crop&w=800&q=60",
    },
    {
      name: "John Doe",
      avatar:
        "https://images.unsplash.com/photo-1696452044585-c6a9389d0c6b?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHx0b3BpYy1mZWVkfDM3fHRvd0paRnNrcEdnfHxlbnwwfHx8fHw%3D&auto=format&fit=crop&w=800&q=60",
    },

    {
      name: "Doubting Thomas",
    },
  ];
  return (
    <Popover.Content align={align} className="w-72">
      <Command>
        <Command.Input autoFocus placeholder={placeholder} />
        <Divider className="my-2" />
        <Command.Empty className="py-2">
          <Text color="muted">No user found.</Text>
        </Command.Empty>
        <Command.Group>
          {users.map(({ name, avatar }, idx) => (
            <Command.Item
              active={idx === 1}
              className="justify-between"
              key={name}
              onSelect={() => {
                onAssigneeSelected(name);
                setOpen(false);
              }}
            >
              <Flex align="center" gap={2}>
                <Avatar color="primary" name={name} size="sm" src={avatar} />
                <Text className="max-w-[10rem] truncate">{name}</Text>
              </Flex>
              <Flex align="center" gap={1}>
                {idx === 1 && (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                )}
                <Text color="muted">{idx}</Text>
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
