"use client";

import { Box, Flex, Text, Button, Tabs } from "ui";
import { ClockIcon, TeamIcon } from "icons";
import { useState } from "react";
import { useMembers } from "@/lib/hooks/members";
import { InviteMembersDialog } from "@/components/ui";
import { usePendingInvitations } from "@/modules/invitations/hooks/pending-invitations";
import { SectionHeader } from "../../components";
import { WorkspaceMember } from "./components/member";

const MembersTab = ({
  setIsInviteMembersDialogOpen,
}: {
  setIsInviteMembersDialogOpen: (open: boolean) => void;
}) => {
  const { data: members = [] } = useMembers();
  return (
    <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
      <SectionHeader
        action={
          <Button
            color="tertiary"
            leftIcon={<TeamIcon />}
            onClick={() => {
              setIsInviteMembersDialogOpen(true);
            }}
            variant="outline"
          >
            Invite Members
          </Button>
        }
        description="Manage members of your workspace and their roles."
        title="Workspace members"
      />

      <Flex
        className="divide-y divide-gray-100 dark:divide-dark-100"
        direction="column"
      >
        {members.map((member) => (
          <WorkspaceMember key={member.id} {...member} />
        ))}
      </Flex>
    </Box>
  );
};

export const WorkspaceMembersSettings = () => {
  const { data: pendingInvitations = [] } = usePendingInvitations();
  const [isInviteMembersDialogOpen, setIsInviteMembersDialogOpen] =
    useState(false);
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Members
      </Text>
      {pendingInvitations.length > 0 ? (
        <Tabs defaultValue="members">
          <Tabs.List className="mx-0 mb-4">
            <Tabs.Tab leftIcon={<TeamIcon />} value="members">
              Members
            </Tabs.Tab>
            <Tabs.Tab leftIcon={<ClockIcon />} value="pending">
              Pending invitations
            </Tabs.Tab>
          </Tabs.List>
          <Tabs.Panel value="members">
            <MembersTab
              setIsInviteMembersDialogOpen={setIsInviteMembersDialogOpen}
            />
          </Tabs.Panel>
          <Tabs.Panel value="pending">Pending invitations</Tabs.Panel>
        </Tabs>
      ) : (
        <MembersTab
          setIsInviteMembersDialogOpen={setIsInviteMembersDialogOpen}
        />
      )}

      <InviteMembersDialog
        isOpen={isInviteMembersDialogOpen}
        setIsOpen={setIsInviteMembersDialogOpen}
      />
    </Box>
  );
};
