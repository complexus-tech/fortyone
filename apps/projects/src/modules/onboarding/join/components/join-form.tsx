"use client";

import { Button, Box, Flex, Text, Wrapper, Avatar } from "ui";
import { toast } from "sonner";
import { useRef, useState } from "react";
import { redirect } from "next/navigation";
import { acceptInvitation } from "@/modules/invitations/public/onboarding";
import type { Invitation } from "@/modules/invitations/public/types";
import { buildWorkspaceUrl } from "@/utils";
import { useWorkspaces } from "@/lib/hooks/workspaces";

export const JoinForm = ({
  invitation,
  token,
}: {
  invitation: Invitation;
  token: string;
}) => {
  const { workspaceName, workspaceSlug } = invitation;
  const [isLoading, setIsLoading] = useState(false);
  const latestRequestVersionRef = useRef(0);
  const { data: workspaces = [] } = useWorkspaces();

  const handleJoin = () => {
    const requestVersion = latestRequestVersionRef.current + 1;
    latestRequestVersionRef.current = requestVersion;
    setIsLoading(true);
    void acceptInvitation(token)
      .then((res) => {
        if (requestVersion !== latestRequestVersionRef.current) {
          return;
        }

        if (res.error?.message) {
          toast.error("Failed to join workspace", {
            description: res.error.message,
          });
          return;
        }

        if (workspaces.length === 0) {
          redirect("/onboarding/account");
          return;
        }

        redirect(buildWorkspaceUrl(workspaceSlug));
      })
      .finally(() => {
        if (requestVersion === latestRequestVersionRef.current) {
          setIsLoading(false);
        }
      });
  };

  return (
    <Box className="space-y-5">
      <Wrapper className="py-3">
        <Flex align="center" gap={3} justify="between">
          <Flex align="center" gap={2}>
            <Avatar
              className="bg-dark dark:text-foreground text-white dark:bg-white"
              color="naked"
              name={workspaceName}
              rounded="md"
            />
            <Box>
              <Text>{workspaceName}</Text>
              <Text color="muted" fontSize="sm">
                {workspaceSlug}.fortyone.app
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
            Accept invitation
          </Button>
        </Flex>
      </Wrapper>
      <Flex align="center" className="my-3 gap-4" justify="between">
        <Box className="bg-surface-muted h-px w-full" />
        <Text className="text-[0.95rem]" color="muted">
          OR
        </Text>
        <Box className="bg-surface-muted h-px w-full" />
      </Flex>
      <Button
        align="center"
        className="opacity-80"
        color="tertiary"
        fullWidth
        href="/onboarding/create"
        size="lg"
        variant="naked"
      >
        Create your own workspace
      </Button>
    </Box>
  );
};
