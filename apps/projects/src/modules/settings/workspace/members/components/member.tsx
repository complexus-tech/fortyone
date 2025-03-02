import { Box, Flex, Text, Button, Avatar, Select, Menu } from "ui";
import { DeleteIcon, MoreHorizontalIcon } from "icons";
import { useState } from "react";
import { useSession } from "next-auth/react";
import { cn } from "lib";
import type { Member, UserRole } from "@/types";
import { ConfirmDialog, RowWrapper } from "@/components/ui";
import { useUpdateRoleMutation } from "@/lib/hooks/update-role";
import { useRemoveMemberMutation } from "@/lib/hooks/remove-member-mutation";
import { useMembers } from "@/lib/hooks/members";

export const WorkspaceMember = ({
  id,
  fullName,
  username,
  avatarUrl,
  role,
  email,
}: Member) => {
  const { data: members = [] } = useMembers();
  const { data: session } = useSession();
  const { mutate: updateRole } = useUpdateRoleMutation();
  const { mutate: removeMember } = useRemoveMemberMutation();
  const [isOpen, setIsOpen] = useState(false);
  const isCurrentUser = session?.user?.id === id;

  const otherAdminsCount = members.filter(
    (member) => member.role === "admin" && member.id !== session?.user?.id,
  ).length;

  const canLeaveWorkspace = otherAdminsCount > 0;

  const handleRemoveUser = () => {
    removeMember(id);
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
        <Select
          disabled={isCurrentUser}
          onValueChange={(value) => {
            updateRole({ userId: id, role: value as UserRole });
          }}
          value={role}
        >
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
              {isCurrentUser ? (
                <Menu.Item
                  disabled={!canLeaveWorkspace}
                  onSelect={() => {
                    if (canLeaveWorkspace) {
                      setIsOpen(true);
                    }
                  }}
                >
                  <DeleteIcon className="h-[1.15rem]" />
                  {canLeaveWorkspace
                    ? "Leave workspace..."
                    : "Cannot leave (no other admin)"}
                </Menu.Item>
              ) : (
                <Menu.Item
                  onSelect={() => {
                    setIsOpen(true);
                  }}
                >
                  <DeleteIcon className="h-[1.15rem]" />
                  Remove user...
                </Menu.Item>
              )}
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
