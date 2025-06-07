"use client";

import { Box, Flex, Text, Button, Tabs } from "ui";
import { ClockIcon, InviteMembersIcon, TeamIcon } from "icons";
import { useState } from "react";
import { useMembers } from "@/lib/hooks/members";
import { InviteMembersDialog } from "@/components/ui";
import { usePendingInvitations } from "@/modules/invitations/hooks/pending-invitations";
import { SectionHeader } from "../../components";
import { WorkspaceMember } from "./components/member";
import { WorkspaceInvitee } from "./components/invitee";

const MembersTab = ({
  setIsInviteMembersDialogOpen,
}: {
  setIsInviteMembersDialogOpen: (open: boolean) => void;
}) => {
  const { data: allMembers = [] } = useMembers();
  const members = allMembers.filter(({ role }) => role !== "system");
  return (
    <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
      <SectionHeader
        action={
          <Button
            color="tertiary"
            leftIcon={<InviteMembersIcon />}
            onClick={() => {
              setIsInviteMembersDialogOpen(true);
            }}
            variant="outline"
          >
            Invite <span className="hidden md:inline">Members</span>
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

const PendingInvitationsTab = ({
  setIsInviteMembersDialogOpen,
}: {
  setIsInviteMembersDialogOpen: (open: boolean) => void;
}) => {
  const { data: pendingInvitations = [] } = usePendingInvitations();
  return (
    <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
      <SectionHeader
        action={
          <Button
            className="shrink-0"
            color="tertiary"
            leftIcon={<InviteMembersIcon />}
            onClick={() => {
              setIsInviteMembersDialogOpen(true);
            }}
            variant="outline"
          >
            Invite More
          </Button>
        }
        description="Pending invitations to join your workspace."
        title="Pending invitations"
      />

      <Flex
        className="divide-y divide-gray-100 dark:divide-dark-100"
        direction="column"
      >
        {pendingInvitations.map((invitation) => (
          <WorkspaceInvitee key={invitation.id} {...invitation} />
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
          <Tabs.List className="mx-0 mb-4 md:mx-0">
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
          <Tabs.Panel value="pending">
            <PendingInvitationsTab
              setIsInviteMembersDialogOpen={setIsInviteMembersDialogOpen}
            />
          </Tabs.Panel>
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
