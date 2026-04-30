"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { CheckIcon, CloseIcon, GitHubIcon, NewTabIcon } from "icons";
import { Badge, Box, Button, Container, Divider, Flex, Text, TimeAgo } from "ui";
import { ConfirmDialog } from "@/components/ui";
import { useWorkspacePath } from "@/hooks";
import { useAcceptIntegrationRequest } from "./hooks/use-accept-request";
import { useDeclineIntegrationRequest } from "./hooks/use-decline-request";
import { useIntegrationRequest } from "./hooks/use-request";

const metadataText = (value: unknown) =>
  typeof value === "string" && value.trim() ? value : null;

export const IntegrationRequestDetails = ({
  requestId,
}: {
  requestId: string;
}) => {
  const { teamId } = useParams<{ teamId: string }>();
  const router = useRouter();
  const { withWorkspace } = useWorkspacePath();
  const [isDeclining, setIsDeclining] = useState(false);
  const { data: request, isPending } = useIntegrationRequest(requestId);
  const acceptRequest = useAcceptIntegrationRequest();
  const declineRequest = useDeclineIntegrationRequest();

  if (isPending) {
    return (
      <Box className="h-dvh px-8 py-7">
        <Box className="bg-surface-muted mb-8 h-8 w-2/5 rounded" />
        <Box className="bg-surface-muted mb-4 h-4 w-4/5 rounded" />
        <Box className="bg-surface-muted h-4 w-3/5 rounded" />
      </Box>
    );
  }

  if (!request) {
    return (
      <Box className="flex h-dvh items-center justify-center px-6">
        <Box>
          <Text align="center" className="mb-3" fontSize="xl">
            Request not found
          </Text>
          <Text align="center" color="muted">
            This request may have already been handled.
          </Text>
        </Box>
      </Box>
    );
  }

  const repositoryName = metadataText(request.metadata.repository_full_name);
  const senderLogin = metadataText(request.metadata.sender_login);
  const issueNumber = request.sourceNumber ? `#${request.sourceNumber}` : "";

  return (
    <Box className="h-dvh overflow-y-auto">
      <Container className="max-w-5xl py-6">
        <Flex align="center" className="mb-8" gap={3} justify="between">
          <Flex align="center" className="min-w-0" gap={2}>
            {request.provider === "github" ? (
              <GitHubIcon className="text-primary h-5 shrink-0" />
            ) : null}
            <Text className="line-clamp-1 font-medium" color="muted">
              {repositoryName ?? "Integration"} {issueNumber}
            </Text>
            <Badge color="tertiary" size="sm" variant="outline">
              {request.status}
            </Badge>
          </Flex>
          {request.sourceUrl ? (
            <Button
              color="tertiary"
              href={request.sourceUrl}
              rightIcon={<NewTabIcon className="h-4" />}
              size="sm"
              target="_blank"
            >
              Open source
            </Button>
          ) : null}
        </Flex>

        <Box className="mb-7">
          <Text className="mb-2 text-3xl font-semibold md:text-4xl">
            {request.title}
          </Text>
          <Flex align="center" gap={2}>
            <Text color="muted">
              Created <TimeAgo timestamp={request.createdAt} />
            </Text>
            {senderLogin ? (
              <Text color="muted">by {senderLogin}</Text>
            ) : null}
          </Flex>
        </Box>

        <Box className="min-h-28">
          {request.description ? (
            <Text className="whitespace-pre-wrap text-lg leading-8">
              {request.description}
            </Text>
          ) : (
            <Text className="text-lg" color="muted">
              No description.
            </Text>
          )}
        </Box>

        <Divider className="my-8" />

        <Flex align="center" gap={2}>
          <Button
            disabled={request.status !== "pending"}
            leftIcon={<CheckIcon className="h-4" />}
            loading={acceptRequest.isPending}
            loadingText="Accepting..."
            onClick={() => {
              acceptRequest.mutate(request.id, {
                onSuccess: (res) => {
                  if (res.data?.acceptedStoryId) {
                    router.push(withWorkspace(`/story/${res.data.acceptedStoryId}`));
                  }
                },
              });
            }}
          >
            Accept
          </Button>
          <Button
            color="tertiary"
            disabled={request.status !== "pending"}
            leftIcon={<CloseIcon className="h-4" />}
            onClick={() => {
              setIsDeclining(true);
            }}
          >
            Decline
          </Button>
        </Flex>
      </Container>

      <ConfirmDialog
        confirmText="Decline request"
        description="Declining removes this item from the team request queue. You can still find the original item in the source integration."
        isLoading={declineRequest.isPending}
        isOpen={isDeclining}
        loadingText="Declining..."
        onCancel={() => {
          setIsDeclining(false);
        }}
        onClose={() => {
          setIsDeclining(false);
        }}
        onConfirm={() => {
          declineRequest.mutate(request.id, {
            onSuccess: (res) => {
              if (!res.error?.message) {
                setIsDeclining(false);
                router.push(withWorkspace(`/teams/${teamId}/requests`));
              }
            },
          });
        }}
        title="Decline this request?"
      />
    </Box>
  );
};
