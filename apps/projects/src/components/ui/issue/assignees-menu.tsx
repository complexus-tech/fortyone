import { CheckIcon } from "icons";
import type { ReactNode } from "react";
import { Avatar, Flex, Menu, Text } from "ui";

export const AssigneesMenu = ({ children }: { children: ReactNode }) => {
  return <Menu>{children}</Menu>;
};

const Trigger = ({ children }: { children: ReactNode }) => (
  <Menu.Button asChild>{children}</Menu.Button>
);

const Items = ({
  isSearchEnabled = true,
  placeholder = "Assign user...",
}: {
  isSearchEnabled?: boolean;
  placeholder?: string;
}) => {
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
    <Menu.Items className="w-72">
      {isSearchEnabled ? (
        <>
          <Menu.Group className="px-4">
            <Menu.Input autoFocus placeholder={placeholder} />
          </Menu.Group>
          <Menu.Separator className="my-2" />
        </>
      ) : null}

      <Menu.Group>
        {users.map(({ name, avatar }, idx) => (
          <Menu.Item active={idx === 1} className="justify-between" key={name}>
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
          </Menu.Item>
        ))}
      </Menu.Group>
    </Menu.Items>
  );
};

AssigneesMenu.Trigger = Trigger;
AssigneesMenu.Items = Items;
