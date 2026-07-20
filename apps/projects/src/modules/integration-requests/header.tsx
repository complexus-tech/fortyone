"use client";

import { useState } from "react";
import { Button, Flex, Menu, Text } from "ui";
import { CheckIcon, CloseIcon, IntakeIcon, MoreVerticalIcon } from "icons";
import { ConfirmDialog } from "@/components/ui";
import { MobileMenuButton } from "@/components/shared";
import { useAcceptAllIntegrationRequests } from "./hooks/use-accept-all-requests";
import { useDeclineAllIntegrationRequests } from "./hooks/use-decline-all-requests";

export const IntegrationRequestsHeader = ({
  requestCount,
  teamId,
}: {
  requestCount: number;
  teamId: string;
}) => {
  const [isAcceptingAll, setIsAcceptingAll] = useState(false);
  const [isDecliningAll, setIsDecliningAll] = useState(false);
  const acceptAllRequests = useAcceptAllIntegrationRequests();
  const declineAllRequests = useDeclineAllIntegrationRequests();
  const hasRequests = requestCount > 0;

  return (
    <Flex
      align="center"
      className="border-border/60 d h-16 border-b-[0.5px] px-4"
      justify="between"
    >
      <Flex align="center" className="gap-2">
        <MobileMenuButton />
        <IntakeIcon className="h-5 w-auto" />
        <Text>Intake</Text>
      </Flex>
      <Flex align="center" gap={2}>
        <Menu>
          <Menu.Button>
            <Button
              asIcon
              color="tertiary"
              disabled={!hasRequests}
              rightIcon={<MoreVerticalIcon />}
              size="sm"
            >
              <span className="sr-only">More options</span>
            </Button>
          </Menu.Button>
          <Menu.Items align="end">
            <Menu.Group className="mt-1 mb-3 px-4">
              <Text color="muted" textOverflow="truncate">
                Manage intake
              </Text>
            </Menu.Group>
            <Menu.Separator className="mb-1.5" />
            <Menu.Group>
              <Menu.Item
                disabled={!hasRequests}
                onSelect={() => {
                  setIsAcceptingAll(true);
                }}
              >
                <CheckIcon className="h-5 w-auto" />
                Accept all intake items
              </Menu.Item>
              <Menu.Item
                className="text-danger"
                disabled={!hasRequests}
                onSelect={() => {
                  setIsDecliningAll(true);
                }}
              >
                <CloseIcon className="text-danger h-5 w-auto" />
                Decline all intake items
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>

      <ConfirmDialog
        confirmText="Accept all intake items"
        description="Accept every pending intake item in this team. Each accepted item will become a story."
        isLoading={acceptAllRequests.isPending}
        isOpen={isAcceptingAll}
        loadingText="Accepting..."
        onClose={() => {
          setIsAcceptingAll(false);
        }}
        onConfirm={() => {
          acceptAllRequests.mutate(teamId, {
            onSuccess: () => {
              setIsAcceptingAll(false);
            },
          });
        }}
        title="Accept all intake items?"
      />

      <ConfirmDialog
        confirmText="Decline all intake items"
        description="Decline every pending intake item in this team. Original items remain available in their source integrations when supported."
        isLoading={declineAllRequests.isPending}
        isOpen={isDecliningAll}
        loadingText="Declining..."
        onClose={() => {
          setIsDecliningAll(false);
        }}
        onConfirm={() => {
          declineAllRequests.mutate(teamId, {
            onSuccess: () => {
              setIsDecliningAll(false);
            },
          });
        }}
        title="Decline all intake items?"
      />
    </Flex>
  );
};
