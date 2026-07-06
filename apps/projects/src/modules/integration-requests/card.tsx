"use client";

import { cn } from "lib";
import { ChatIcon, CheckIcon, CloseIcon, GitHubIcon, SlackIcon } from "icons";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";
import { Box, ContextMenu, Flex, Text, TimeAgo } from "ui";
import { ConfirmDialog, PriorityIcon } from "@/components/ui";
import { useWorkspacePath } from "@/hooks";
import type { IntegrationRequest } from "./types";
import { useAcceptIntegrationRequest } from "./hooks/use-accept-request";
import { useDeclineIntegrationRequest } from "./hooks/use-decline-request";

const providerLabel = (provider: IntegrationRequest["provider"]) => {
  switch (provider) {
    case "github":
      return "GitHub";
    case "slack":
      return "Slack";
    case "intercom":
      return "Intercom";
    default:
      return "Integration";
  }
};

const providerIcon = (provider: IntegrationRequest["provider"]) => {
  switch (provider) {
    case "github":
      return <GitHubIcon className="h-4 shrink-0" />;
    case "slack":
      return <SlackIcon className="h-4 shrink-0" />;
    case "intercom":
      return <ChatIcon className="h-4 shrink-0" />;
    default:
      return null;
  }
};

export const IntegrationRequestCard = ({
  request,
  index,
}: {
  request: IntegrationRequest;
  index: number;
}) => {
  const pathname = usePathname();
  const { withWorkspace } = useWorkspacePath();
  const acceptRequest = useAcceptIntegrationRequest();
  const declineRequest = useDeclineIntegrationRequest();
  const [isDeclining, setIsDeclining] = useState(false);
  const sourceNumber = request.sourceNumber ? `#${request.sourceNumber}` : "";
  const isActive = pathname.includes(request.id);

  return (
    <ContextMenu>
      <ContextMenu.Trigger>
        <Link
          className="block"
          href={withWorkspace(
            `/teams/${request.teamId}/requests/${request.id}`,
          )}
          prefetch={index <= 10 ? true : null}
        >
          <Box
            className={cn(
              "border-border hover:bg-surface-muted d d block cursor-pointer border-b-[0.5px] px-5 py-[0.655rem] transition md:px-4",
              {
                "bg-surface-muted": isActive,
              },
            )}
          >
            <Flex align="center" className="mb-2" gap={2} justify="between">
              <Text className="line-clamp-1 flex-1 font-medium">
                {request.title}
              </Text>
              <Text className="shrink-0 text-[0.95rem]" color="muted">
                <TimeAgo timestamp={request.createdAt} />
              </Text>
            </Flex>
            <Flex align="center" gap={3} justify="between">
              <Flex align="center" className="min-w-0 flex-1" gap={2}>
                {providerIcon(request.provider)}
                <Text className="line-clamp-1" color="muted">
                  {providerLabel(request.provider)} {request.sourceType}{" "}
                  {sourceNumber}
                </Text>
              </Flex>
              <PriorityIcon className="shrink-0" priority={request.priority} />
            </Flex>
          </Box>
        </Link>
      </ContextMenu.Trigger>
      <ContextMenu.Items>
        <ContextMenu.Group>
          <ContextMenu.Item
            disabled={request.status !== "pending"}
            onSelect={() => {
              acceptRequest.mutate(request.id);
            }}
          >
            <CheckIcon />
            Accept
          </ContextMenu.Item>
          <ContextMenu.Item
            className="text-danger"
            disabled={request.status !== "pending"}
            onSelect={() => {
              setIsDeclining(true);
            }}
          >
            <CloseIcon className="text-danger" />
            Decline
          </ContextMenu.Item>
        </ContextMenu.Group>
      </ContextMenu.Items>
      <ConfirmDialog
        confirmText="Decline request"
        description="Declining removes this item from the team request queue."
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
            onSuccess: () => {
              setIsDeclining(false);
            },
          });
        }}
        title="Decline this request?"
      />
    </ContextMenu>
  );
};
