"use client";

import { Button, Box, Flex, Text, Wrapper, Avatar } from "ui";
import { toast } from "sonner";
import { useState } from "react";
import { useSearchParams } from "next/navigation";
import type { Invitation } from "@/lib/actions/verify-invitation";
import { acceptInvitation } from "@/lib/actions/accept-invitation";

export const CreateAccountForm = ({
  invitation,
}: {
  invitation: Invitation;
}) => {
  const searchParams = useSearchParams();
  const token = searchParams.get("token");
  const [isLoading, setIsLoading] = useState(false);

  const handleJoin = async () => {
    setIsLoading(true);
    const res = await acceptInvitation(token || "");
    if (res.error?.message) {
      toast.error("Failed to join workspace", {
        description: res.error.message,
      });
      setIsLoading(false);
    }
  };

  return (
    <Box className="space-y-5">
      <Wrapper className="py-3">
        <Flex align="center" gap={3} justify="between">
          <Flex align="center" gap={2}>
            <Avatar
              name={invitation.workspaceName}
              rounded="md"
              style={{ backgroundColor: invitation.workspaceColor }}
            />
            <Box>
              <Text>{invitation.workspaceName}</Text>
              <Text color="muted" fontSize="sm">
                {invitation.workspaceSlug}.complexus.app
              </Text>
            </Box>
          </Flex>

          <Button
            color="tertiary"
            loading={isLoading}
            loadingText="Joining..."
            onClick={handleJoin}
            size="sm"
          >
            Join now
          </Button>
        </Flex>
      </Wrapper>
      <Flex align="center" className="my-3 gap-4" justify="between">
        <Box className="h-px w-full bg-white/10" />
        <Text className="text-[0.95rem] opacity-40" color="white">
          OR
        </Text>
        <Box className="h-px w-full bg-white/10" />
      </Flex>
      <Button
        align="center"
        className="opacity-80"
        color="tertiary"
        fullWidth
        href="/onboarding/create"
        variant="naked"
      >
        Create your own workspace
      </Button>
    </Box>
  );
};
