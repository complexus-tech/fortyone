import {
  Box,
  Flex,
  Text,
  Button,
  Avatar,
  Select,
  Menu,
  Dialog,
  Input,
} from "ui";
import { DeleteIcon, EditIcon, MoreHorizontalIcon } from "icons";
import { useState } from "react";
import { useSession } from "next-auth/react";
import type { Member } from "@/types";
import { RowWrapper } from "@/components/ui";

export const WorkspaceMember = ({
  id,
  fullName,
  username,
  avatarUrl,
}: Member) => {
  const { data: session } = useSession();
  const [isOpen, setIsOpen] = useState(false);
  const isCurrentUser = session?.user?.id === id;
  return (
    <RowWrapper className="border-0 px-6 py-3">
      <Flex align="center" gap={3}>
        <Avatar name={fullName} src={avatarUrl} />
        <Box>
          <Text className="font-medium">{fullName}</Text>
          <Text color="muted">{username}</Text>
        </Box>
      </Flex>
      <Flex align="center" gap={3}>
        {!isCurrentUser ? (
          <Select defaultValue="member">
            <Select.Trigger className="w-32">
              <Select.Input />
            </Select.Trigger>
            <Select.Content>
              <Select.Option value="admin">Admin</Select.Option>
              <Select.Option value="member">Member</Select.Option>
              <Select.Option value="guest">Guest</Select.Option>
            </Select.Content>
          </Select>
        ) : null}
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
                <EditIcon />
                Edit...
              </Menu.Item>
            </Menu.Group>
            <Menu.Separator />
            <Menu.Group>
              <Menu.Item>
                <DeleteIcon className="h-[1.15rem]" />
                Remove user...
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>

      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content>
          <Dialog.Header>
            <Dialog.Title className="px-6 text-lg">
              Edit Member Details
            </Dialog.Title>
            <Dialog.Description>
              Make changes to the member&apos;s information here.
            </Dialog.Description>
          </Dialog.Header>
          <Dialog.Body className="flex flex-col gap-4">
            <Input defaultValue={fullName} id="fullName" label="Full Name" />
            <Input defaultValue={username} id="username" label="Username" />
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-3 border-0 pt-1">
            <Button color="tertiary">Cancel</Button>
            <Button>Save changes</Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </RowWrapper>
  );
};
