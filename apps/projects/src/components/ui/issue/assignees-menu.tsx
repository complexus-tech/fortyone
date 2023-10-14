import { TbCheck } from "react-icons/tb";
import { Avatar, Button, Flex, Menu, Text } from "ui";

export const AssigneesMenu = ({
  isSearchEnabled = false,
}: {
  isSearchEnabled?: boolean;
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
          className="select-none px-1"
          color="tertiary"
          leftIcon={
            <Avatar
              color="gray"
              name="Joseph Mukorivo"
              size="sm"
              src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
            />
          }
          size="sm"
          variant="naked"
        >
          <span className="sr-only">Assign user</span>
        </Button>
      </Menu.Button>
      <Menu.Items align="end" className="w-72">
        {isSearchEnabled ? (
          <>
            <Menu.Group className="mb-2 px-4">
              <Menu.Input autoFocus placeholder="Assign user" />
            </Menu.Group>
            <Menu.Separator />
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
                  <TbCheck className="h-5 w-auto" strokeWidth={2.1} />
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
