"use client";

import { Box, Button, Text } from "ui";
import { useMyInvitations } from "@/modules/invitations/hooks/my-invitations";
import { SectionHeader } from "../components";
import { InvitationRow } from "./components/invite";

export const InvitationsPage = () => {
  const { data: invitations = [] } = useMyInvitations();

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Workspace Invitations
      </Text>
      <Box className="rounded-[0.6rem] border border-border bg-surface">
        <SectionHeader
          action={
            invitations.length > 1 && (
              <Button color="tertiary">Accept All</Button>
            )
          }
          description="View and manage your pending workspace invitations."
          title="Pending Invitations"
        />
        <Box className="divide-y divide-gray-100 dark:divide-dark-100">
          {invitations.length === 0 ? (
            <Box className="px-6 py-8 text-center">
              <Text className="font-medium">No pending invitations</Text>
              <Text className="mt-2" color="muted">
                When you receive invitations to join workspaces, they will
                appear here
              </Text>
            </Box>
          ) : (
            invitations.map((invitation) => (
              <InvitationRow invitation={invitation} key={invitation.id} />
            ))
          )}
        </Box>
      </Box>
    </Box>
  );
};
