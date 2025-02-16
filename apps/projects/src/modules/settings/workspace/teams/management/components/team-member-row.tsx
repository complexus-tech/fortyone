"use client";

import { useState } from "react";
import { Box, Flex, Text, Button, Avatar, Select, Menu } from "ui";
import { MoreHorizontalIcon, DeleteIcon } from "icons";
import type { Member } from "@/types";
import type { TeamMember } from "@/modules/teams/types";
import { ConfirmDialog, RowWrapper } from "@/components/ui";

type TeamMemberRowProps = {
  member: Member;
  role: "admin" | "member";
  onRemove: () => void;
  teamId: string;
  teamMembers: TeamMember[];
};

export const TeamMemberRow = ({
  member,
  role,
  onRemove,
}: TeamMemberRowProps) => {
  const [isRemoveOpen, setIsRemoveOpen] = useState(false);

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
        <Select defaultValue={role}>
          <Select.Trigger className="w-32">
            <Select.Input />
          </Select.Trigger>
          <Select.Content>
            <Select.Option value="admin">Admin</Select.Option>
            <Select.Option value="member">Member</Select.Option>
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
                  setIsRemoveOpen(true);
                }}
              >
                <DeleteIcon className="h-[1.15rem]" />
                Remove member
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>

      <ConfirmDialog
        confirmText="Remove member"
        description={`Are you sure you want to remove ${member.fullName} from this team?`}
        isOpen={isRemoveOpen}
        onClose={() => {
          setIsRemoveOpen(false);
        }}
        onConfirm={() => {
          onRemove();
          setIsRemoveOpen(false);
        }}
        title="Remove team member"
      />
    </RowWrapper>
  );
};
