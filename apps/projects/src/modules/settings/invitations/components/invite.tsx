"use client";
import { Box, Text, Flex, Button, Avatar } from "ui";
import { formatDistanceToNow } from "date-fns";
import { useState } from "react";
import { ConfirmDialog, RowWrapper } from "@/components/ui";
import type { Invitation } from "@/modules/invitations/types";
import { useRevokeInvitationMutation } from "@/modules/invitations/hooks/use-revoke-invitation";
import { useAcceptInvitationMutation } from "@/modules/invitations/hooks/use-accept-invitation";

export const InvitationRow = ({ invitation }: { invitation: Invitation }) => {
  const { mutate: declineInvitation } = useRevokeInvitationMutation();
  const { mutate: acceptInvitation } = useAcceptInvitationMutation();

  const [isOpen, setIsOpen] = useState(false);
  const timeLeft = formatDistanceToNow(new Date(invitation.expiresAt), {
    addSuffix: true,
  });

  const handleAccept = () => {
    acceptInvitation(invitation.token!);
  };

  const handleDecline = () => {
    declineInvitation(invitation.id);
    setIsOpen(false);
  };

  return (
    <RowWrapper
      className="w-full px-6 py-4 last-of-type:border-b-0 md:px-6"
      key={invitation.id}
    >
      <Flex align="center" gap={2}>
        <Avatar
          name={invitation.workspaceName}
          rounded="md"
          style={{
            backgroundColor: invitation.workspaceColor,
          }}
        />
        <Box>
          <Text>
            {invitation.workspaceName}{" "}
            <Text as="span" color="muted">
              (expires {timeLeft})
            </Text>
          </Text>
          <Text color="muted" fontSize="sm">
            {invitation.workspaceSlug} &bull; {invitation.role}
          </Text>
        </Box>
      </Flex>
      <Flex gap={2}>
        <Button
          color="primary"
          onClick={handleAccept}
          size="sm"
          variant="naked"
        >
          Accept & Join
        </Button>
        <Button
          color="tertiary"
          onClick={() => {
            setIsOpen(true);
          }}
          size="sm"
          variant="outline"
        >
          Decline
        </Button>
      </Flex>

      <ConfirmDialog
        confirmText="Yes, decline"
        description="Are you sure you want to decline this invitation? You will not be able to join this workspace until you get a new invitation."
        isOpen={isOpen}
        onClose={() => {
          setIsOpen(false);
        }}
        onConfirm={handleDecline}
        title="Decline Invitation?"
      />
    </RowWrapper>
  );
};
