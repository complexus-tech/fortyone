"use client";

import { useState } from "react";
import { Box, Flex, Text, Button, Avatar, Menu } from "ui";
import { MoreHorizontalIcon, DeleteIcon } from "icons";
import { useSession } from "next-auth/react";
import type { Member } from "@/types";
import { ConfirmDialog, RowWrapper } from "@/components/ui";
import { useUserRole } from "@/hooks/role";
import { useRemoveMemberMutation } from "@/modules/teams/hooks/remove-member-mutation";

type TeamMemberRowProps = {
  member: Member;
  teamId: string;
};

export const TeamMemberRow = ({ member, teamId }: TeamMemberRowProps) => {
  const { data: session } = useSession();
  const { userRole } = useUserRole();
  const { mutate: removeMember } = useRemoveMemberMutation();
  const [isRemoveOpen, setIsRemoveOpen] = useState(false);
  const isCurrentUser = session?.user?.id === member.id;

  const handleRemoveMember = () => {
    removeMember({ teamId, memberId: member.id });
    setIsRemoveOpen(false);
  };

  return (
    <RowWrapper className="px-6 py-4">
      <Flex align="center" gap={3}>
        <Avatar name={member.fullName} src={member.avatarUrl} />
        <Box>
          <Text className="font-medium">
            {member.fullName}
            <Text as="span" color="muted">
              ({member.username})
            </Text>
          </Text>
          <Text color="muted">{member.email}</Text>
        </Box>
      </Flex>

      <Flex align="center" gap={3}>
        <Text className="w-20 first-letter:uppercase" color="muted">
          {member.role}
        </Text>
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
                  setIsRemoveOpen(true);
                }}
              >
                <DeleteIcon className="h-[1.15rem]" />
                {isCurrentUser ? "Leave team" : "Remove member"}
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>

      <ConfirmDialog
        confirmText={isCurrentUser ? "Leave team" : "Remove member"}
        description={
          isCurrentUser
            ? `Are you sure you want to leave this team? ${userRole === "admin" ? "You can always join again later." : ""}`
            : `Are you sure you want to remove ${member.fullName} from this team?`
        }
        isOpen={isRemoveOpen}
        onClose={() => {
          setIsRemoveOpen(false);
        }}
        onConfirm={handleRemoveMember}
        title={isCurrentUser ? "Leave team" : "Remove member"}
      />
    </RowWrapper>
  );
};
