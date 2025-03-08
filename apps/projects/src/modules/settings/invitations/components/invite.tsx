import { Box, Text, Flex, Button, Avatar } from "ui";
import { formatDistanceToNow } from "date-fns";
import { RowWrapper } from "@/components/ui";
import type { Invitation } from "@/modules/invitations/types";

const domain = process.env.NEXT_PUBLIC_DOMAIN;

export const InvitationRow = ({ invitation }: { invitation: Invitation }) => {
  const timeLeft = formatDistanceToNow(new Date(invitation.expiresAt), {
    addSuffix: true,
  });
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
            {invitation.workspaceSlug}.{domain}
          </Text>
        </Box>
      </Flex>
      <Flex gap={2}>
        <Button
          color="primary"
          loading={false}
          loadingText="Accepting..."
          size="sm"
        >
          Accept
        </Button>
        <Button
          color="tertiary"
          loading={false}
          loadingText="Declining..."
          size="sm"
          variant="outline"
        >
          Decline
        </Button>
      </Flex>
    </RowWrapper>
  );
};
