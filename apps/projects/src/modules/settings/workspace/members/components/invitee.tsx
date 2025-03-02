import { Flex, Text, Button, Menu, Box, Avatar } from "ui";
import { DeleteIcon, MoreHorizontalIcon } from "icons";
import { useState } from "react";
import { formatDistanceToNow } from "date-fns";
import type { Invitation } from "@/modules/invitations/types";
import { ConfirmDialog, RowWrapper } from "@/components/ui";
import { useMembers } from "@/lib/hooks/members";
import { useRevokeInvitationMutation } from "@/modules/invitations/hooks/use-revoke-invitation";

export const WorkspaceInvitee = ({
  id,
  email,
  role,
  inviterId,
  expiresAt,
}: Invitation) => {
  const [isOpen, setIsOpen] = useState(false);
  const { data: members = [] } = useMembers();
  const revokeMutation = useRevokeInvitationMutation();

  const inviter = members.find((member) => member.id === inviterId);
  const timeLeft = formatDistanceToNow(new Date(expiresAt), {
    addSuffix: true,
  });

  const handleRevokeInvitation = async () => {
    await revokeMutation.mutateAsync(id);
    setIsOpen(false);
  };

  return (
    <RowWrapper className="border-0 px-6 py-3">
      <Flex align="center" gap={3}>
        <Avatar
          className="border border-dashed dark:border-dark-50 dark:bg-dark-100"
          color="tertiary"
          name={email}
        />
        <Box>
          <Text className="font-medium">
            {email}{" "}
            <Text as="span" color="muted">
              (expires {timeLeft})
            </Text>
          </Text>
          <Text color="muted">
            Invited by {inviter?.fullName || inviter?.username || "Unknown"} as{" "}
            <span>
              {role === "admin" ? "an " : "a "}
              <span className="inline-block font-semibold first-letter:uppercase">
                {role}
              </span>
            </span>
          </Text>
        </Box>
      </Flex>

      <Flex align="center" gap={3}>
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
                Revoke invitation
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>

      <ConfirmDialog
        confirmText="Revoke invitation"
        description="Are you sure you want to revoke this invitation? This action cannot be undone."
        isOpen={isOpen}
        onClose={() => {
          setIsOpen(false);
        }}
        onConfirm={handleRevokeInvitation}
        title="Revoke invitation"
      />
    </RowWrapper>
  );
};
