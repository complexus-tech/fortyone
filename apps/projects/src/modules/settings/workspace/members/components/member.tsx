import { Box, Flex, Text, Button, Avatar, Select, Menu } from "ui";
import { DeleteIcon, MoreHorizontalIcon } from "icons";
import { useState } from "react";
import { useSession } from "next-auth/react";
import { cn } from "lib";
import type { Member } from "@/types";
import { ConfirmDialog, RowWrapper } from "@/components/ui";

export const WorkspaceMember = ({
  id,
  fullName,
  username,
  avatarUrl,
  email,
}: Member) => {
  const { data: session } = useSession();
  const [isOpen, setIsOpen] = useState(false);
  const isCurrentUser = session?.user?.id === id;

  const handleRemoveUser = async () => {
    // TODO: remove user
  };

  return (
    <RowWrapper className="border-0 px-6 py-3">
      <Flex align="center" gap={3}>
        <Avatar name={fullName} src={avatarUrl} />
        <Box>
          <Text className="font-medium">
            {fullName}{" "}
            <Text as="span" color="muted">
              ({username})
            </Text>
          </Text>
          <Text color="muted">{email}</Text>
        </Box>
      </Flex>
      <Flex align="center" gap={3}>
        <Select defaultValue="member" disabled={isCurrentUser}>
          <Select.Trigger
            className={cn("w-32", {
              "opacity-50": isCurrentUser,
            })}
          >
            <Select.Input />
          </Select.Trigger>
          <Select.Content>
            <Select.Option value="admin">Admin</Select.Option>
            <Select.Option value="member">Member</Select.Option>
            <Select.Option value="guest">Guest</Select.Option>
          </Select.Content>
        </Select>

        <Menu>
          <Menu.Button>
            <Button
              asIcon
              color="tertiary"
              leftIcon={<MoreHorizontalIcon />}
              rounded="full"
              size="sm"
            >
              <span className="sr-only">More options</span>
            </Button>
          </Menu.Button>
          <Menu.Items align="end">
            <Menu.Group>
              <Menu.Item
                onSelect={() => {
                  setIsOpen(true);
                }}
              >
                <DeleteIcon className="h-[1.15rem]" />
                {isCurrentUser ? "Leave workspace..." : "Remove user..."}
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
      <ConfirmDialog
        confirmText={isCurrentUser ? "Leave workspace" : "Remove user"}
        description={
          isCurrentUser
            ? "Are you sure you want to leave this workspace? You will no longer have access to this workspace. This action cannot be undone."
            : "Are you sure you want to remove this user? They will no longer have access to this workspace. This action cannot be undone."
        }
        isOpen={isOpen}
        onClose={() => {
          setIsOpen(false);
        }}
        onConfirm={handleRemoveUser}
        title={isCurrentUser ? "Leave workspace" : "Remove user"}
      />
    </RowWrapper>
  );
};
