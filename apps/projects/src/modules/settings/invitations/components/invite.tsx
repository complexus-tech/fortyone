"use client";
import { Box, Text, Flex, Button, Avatar } from "ui";
import { formatDistanceToNow } from "date-fns";
import { toast } from "sonner";
import { useQueryClient } from "@tanstack/react-query";
import { redirect } from "next/navigation";
import { useState } from "react";
import { ConfirmDialog, RowWrapper } from "@/components/ui";
import type { Invitation } from "@/modules/invitations/types";
import { acceptInvitation } from "@/modules/invitations/actions/accept-invitation";
import { invitationKeys, workspaceKeys } from "@/constants/keys";
import { useRevokeInvitationMutation } from "@/modules/invitations/hooks/use-revoke-invitation";

const domain = process.env.NEXT_PUBLIC_DOMAIN!;

const getRedirectUrl = (slug: string) => {
  if (domain.includes("localhost")) {
    return `http://${slug}.${domain}/my-work`;
  }
  return `https://${slug}.${domain}/my-work`;
};

export const InvitationRow = ({ invitation }: { invitation: Invitation }) => {
  const queryClient = useQueryClient();
  const { mutate: declineInvitation } = useRevokeInvitationMutation();
  const [isLoading, setIsLoading] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const timeLeft = formatDistanceToNow(new Date(invitation.expiresAt), {
    addSuffix: true,
  });

  const handleAccept = async () => {
    const toastId = "accept-invitation";
    setIsLoading(true);
    toast.loading("Accepting...", {
      id: toastId,
      description: "Please wait...",
    });
    const res = await acceptInvitation(invitation.token!);
    if (res.error?.message) {
      toast.error("Failed to accept invitation", {
        id: toastId,
        description:
          res.error.message || "There was an error accepting the invitation",
      });
      setIsLoading(false);
      return;
    }

    await Promise.all([
      queryClient.invalidateQueries({
        queryKey: invitationKeys.mine,
      }),
      queryClient.invalidateQueries({
        queryKey: workspaceKeys.lists(),
      }),
    ]);
    setIsLoading(false);
    toast.success("Invitation accepted", {
      id: toastId,
      description: "You can now access the workspace",
      action: {
        label: "Open",
        onClick: () => {
          redirect(getRedirectUrl(invitation.workspaceSlug));
        },
      },
    });
  };

  const handleDecline = () => {
    declineInvitation(invitation.id);
    setIsOpen(false);
  };

  return (
    <RowWrapper
      className="w-full px-6 py-4 last-of-type:border-b-0"
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
            {invitation.workspaceSlug}.{domain} &bull; {invitation.role}
          </Text>
        </Box>
      </Flex>
      <Flex gap={2}>
        <Button
          color="primary"
          loading={isLoading}
          loadingText="Accepting..."
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
