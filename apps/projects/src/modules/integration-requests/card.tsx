"use client";

import { cn } from "lib";
import { CheckIcon, CloseIcon, GitHubIcon } from "icons";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Box, ContextMenu, Flex, Text, TimeAgo } from "ui";
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
              "border-border hover:bg-surface-muted cursor-pointer border-b-[0.5px] px-5 py-3 transition md:px-4",
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
            <Flex align="center" gap={2} justify="between">
              <Flex align="center" className="min-w-0" gap={2}>
                {request.provider === "github" ? (
                  <GitHubIcon className="text-primary h-4 shrink-0" />
                ) : null}
                <Text className="line-clamp-1" color="muted">
                  {providerLabel(request.provider)} {request.sourceType}{" "}
                  {sourceNumber}
                </Text>
              </Flex>
              <Text className="shrink-0" color="muted">
                {request.priority}
              </Text>
            </Flex>
          </Box>
        </Link>
      </ContextMenu.Trigger>
      <ContextMenu.Items>
        <ContextMenu.Group>
          <ContextMenu.Item
            disabled={request.status !== "pending"}
            onSelect={() => acceptRequest.mutate(request.id)}
          >
            <CheckIcon />
            Accept
          </ContextMenu.Item>
          <ContextMenu.Item
            className="text-danger"
            disabled={request.status !== "pending"}
            onSelect={() => declineRequest.mutate(request.id)}
          >
            <CloseIcon className="text-danger" />
            Decline
          </ContextMenu.Item>
        </ContextMenu.Group>
      </ContextMenu.Items>
    </ContextMenu>
  );
};
