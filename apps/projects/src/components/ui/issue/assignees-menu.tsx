import { Avatar, Button, Flex, Menu, Text } from "ui";
import { cn } from "lib";
import { Check } from "lucide-react";

export const AssigneesMenu = ({
  isSearchEnabled = true,
  asIcon = true,
  user,
  placeholder = "Assign user...",
}: {
  isSearchEnabled?: boolean;
  asIcon?: boolean;
  user?: string;
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
    <Menu>
      <Menu.Button>
        <Button
          className={cn("gap-2 px-2", {
            "select-none px-1": asIcon,
          })}
          color="tertiary"
          leftIcon={
            <Avatar
              className="h-6"
              color="gray"
              name="Joseph Mukorivo"
              size="sm"
              src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
            />
          }
          size={asIcon ? "sm" : "md"}
          variant="naked"
        >
          <span className="sr-only">Assign user</span>
          {asIcon ? null : (
            <span className="relative -top-[1px] block max-w-[9rem] truncate">
              {user}
            </span>
          )}
        </Button>
      </Menu.Button>
      <Menu.Items align="end" className="w-72">
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
            <Menu.Item
              active={idx === 1}
              className="justify-between"
              key={name}
            >
              <Flex align="center" gap={2}>
                <Avatar color="primary" name={name} size="sm" src={avatar} />
                <Text className="max-w-[10rem] truncate">{name}</Text>
              </Flex>
              <Flex align="center" gap={1}>
                {idx === 1 && (
                  <Check className="h-5 w-auto" strokeWidth={2.1} />
                )}
                <Text color="muted">{idx}</Text>
              </Flex>
            </Menu.Item>
          ))}
        </Menu.Group>
      </Menu.Items>
    </Menu>
  );
};
