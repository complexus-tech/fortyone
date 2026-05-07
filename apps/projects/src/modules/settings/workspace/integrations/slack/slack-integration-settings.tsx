"use client";

import Link from "next/link";
import { useState } from "react";
import { Badge, Box, Button, Dialog, Flex, Menu, Text } from "ui";
import { MoreHorizontalIcon, ReloadIcon, UnlinkIcon } from "icons";
import { useWorkspacePath } from "@/hooks";
import { SectionHeader } from "@/modules/settings/components";
import {
  useCreateSlackInstallSession,
  useDisconnectSlackWorkspace,
  useResyncSlackChannels,
  useSlackIntegration,
} from "@/lib/hooks/slack";

const SlackIcon = ({ className = "h-5 w-5" }: { className?: string }) => (
  <svg
    className={className}
    viewBox="0 0 54 54"
    xmlns="http://www.w3.org/2000/svg"
  >
    <path
      d="M19.712.133a5.381 5.381 0 0 0-5.376 5.387 5.381 5.381 0 0 0 5.376 5.386h5.376V5.52A5.381 5.381 0 0 0 19.712.133m0 14.365H5.376A5.381 5.381 0 0 0 0 19.884a5.381 5.381 0 0 0 5.376 5.387h14.336a5.381 5.381 0 0 0 5.376-5.387 5.381 5.381 0 0 0-5.376-5.386"
      fill="#36C5F0"
    />
    <path
      d="M53.76 19.884a5.381 5.381 0 0 0-5.376-5.386 5.381 5.381 0 0 0-5.376 5.386v5.387h5.376a5.381 5.381 0 0 0 5.376-5.387m-14.336 0V5.52A5.381 5.381 0 0 0 34.048.133a5.381 5.381 0 0 0-5.376 5.387v14.364a5.381 5.381 0 0 0 5.376 5.387 5.381 5.381 0 0 0 5.376-5.387"
      fill="#2EB67D"
    />
    <path
      d="M34.048 54a5.381 5.381 0 0 0 5.376-5.387 5.381 5.381 0 0 0-5.376-5.386h-5.376v5.386A5.381 5.381 0 0 0 34.048 54m0-14.365h14.336a5.381 5.381 0 0 0 5.376-5.386 5.381 5.381 0 0 0-5.376-5.387H34.048a5.381 5.381 0 0 0-5.376 5.387 5.381 5.381 0 0 0 5.376 5.386"
      fill="#ECB22E"
    />
    <path
      d="M0 34.249a5.381 5.381 0 0 0 5.376 5.386 5.381 5.381 0 0 0 5.376-5.386v-5.387H5.376A5.381 5.381 0 0 0 0 34.25m14.336 0v14.364A5.381 5.381 0 0 0 19.712 54a5.381 5.381 0 0 0 5.376-5.387V34.25a5.381 5.381 0 0 0-5.376-5.387 5.381 5.381 0 0 0-5.376 5.386"
      fill="#E01E5A"
    />
  </svg>
);

export const SlackIntegrationSettings = () => {
  const { data: integration } = useSlackIntegration();
  const { withWorkspace } = useWorkspacePath();

  const createInstallSession = useCreateSlackInstallSession();
  const disconnectWorkspace = useDisconnectSlackWorkspace();
  const resyncChannels = useResyncSlackChannels();
  const [isDisconnectOpen, setIsDisconnectOpen] = useState(false);

  const isConnected = Boolean(integration?.slackWorkspace?.isActive);
  const slackWorkspace = integration?.slackWorkspace;

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Slack
      </Text>

      <Box className="border-border bg-surface rounded-2xl border">
        <SectionHeader
          action={
            <Flex gap={2}>
              <Button
                color="invert"
                onClick={() => {
                  createInstallSession.mutate();
                }}
              >
                {isConnected ? "Reconnect Slack" : "Connect Slack"}
              </Button>
            </Flex>
          }
          description="Connect Slack so people can create FortyOne tasks and requests from Slack."
          title="Connected workspace"
        />

        {!slackWorkspace ? (
          <Box className="px-6 py-8">
            <Text className="font-medium">No Slack workspace connected</Text>
            <Text className="mt-1" color="muted">
              Connect Slack to create tasks from slash commands and message
              actions.
            </Text>
          </Box>
        ) : (
          <Flex align="center" className="px-6 py-4" justify="between">
            <Flex align="center" gap={3}>
              <Flex
                align="center"
                className="bg-surface-muted size-9 shrink-0 rounded-lg"
                justify="center"
              >
                <SlackIcon className="h-4.5 w-4.5" />
              </Flex>
              <Box>
                <Text className="font-medium">
                  {slackWorkspace.slackTeamName}
                </Text>
                <Text color="muted">{slackWorkspace.slackTeamDomain}</Text>
              </Box>
            </Flex>
            <Flex align="center" gap={2}>
              <Badge
                className="uppercase"
                color={isConnected ? "success" : "tertiary"}
              >
                {isConnected ? "Connected" : "Disconnected"}
              </Badge>
              {isConnected ? (
                <Menu>
                  <Menu.Button>
                    <Button
                      className="px-2"
                      color="tertiary"
                      leftIcon={<MoreHorizontalIcon />}
                    />
                  </Menu.Button>
                  <Menu.Items align="end">
                    <Menu.Group>
                      <Menu.Item
                        onSelect={() => {
                          resyncChannels.mutate();
                        }}
                      >
                        <ReloadIcon />
                        Resync channels
                      </Menu.Item>
                      <Menu.Item
                        onSelect={() => {
                          createInstallSession.mutate();
                        }}
                      >
                        <SlackIcon className="h-4 w-4" />
                        Update connection
                      </Menu.Item>
                      <Menu.Item
                        className="text-danger"
                        onSelect={() => {
                          setIsDisconnectOpen(true);
                        }}
                      >
                        <UnlinkIcon className="text-danger" />
                        Disconnect workspace
                      </Menu.Item>
                    </Menu.Group>
                  </Menu.Items>
                </Menu>
              ) : null}
            </Flex>
          </Flex>
        )}
      </Box>

      <Box className="mt-6">
        <Link href={withWorkspace("/settings/workspace/integrations")}>
          <Text color="muted">Back to integrations</Text>
        </Link>
      </Box>

      <Dialog onOpenChange={setIsDisconnectOpen} open={isDisconnectOpen}>
        <Dialog.Content>
          <Dialog.Header>
            <Dialog.Title className="px-6 pt-0.5 text-lg">
              Disconnect Slack workspace
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Text color="muted">
              Slash commands and message actions from this Slack workspace will
              stop creating FortyOne tasks and requests.
            </Text>
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-3 border-0 pt-2">
            <Button
              className="px-4"
              color="tertiary"
              onClick={() => {
                setIsDisconnectOpen(false);
              }}
            >
              Cancel
            </Button>
            <Button
              className="px-4"
              loading={disconnectWorkspace.isPending}
              onClick={() => {
                disconnectWorkspace.mutate(undefined, {
                  onSuccess: (res) => {
                    if (!res.error) {
                      setIsDisconnectOpen(false);
                    }
                  },
                });
              }}
            >
              Disconnect
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </Box>
  );
};
