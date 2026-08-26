"use client";

import { useState } from "react";
import { format } from "date-fns";
import { toast } from "sonner";
import { Box, Button, Dialog, Flex, Text } from "ui";
import { SlackIcon, UnlinkIcon } from "icons";
import type { SlackAccountLinkStatus } from "@/lib/hooks/slack/use-account-link-token";
import { useWorkspacePath } from "@/hooks";
import {
  useCreateSlackAccountLinkSession,
  useDisconnectSlackAccount,
  useSlackAccountLinkToken,
  useSlackIntegration,
} from "@/lib/hooks/slack";
import { SettingsBackButton } from "@/modules/settings/components";

const getAccountLinkCopy = ({
  errorMessage,
  hasToken,
  isLinked,
  status,
}: {
  errorMessage: string | null;
  hasToken: boolean;
  isLinked: boolean;
  status: SlackAccountLinkStatus;
}) => {
  if (isLinked || status === "already_connected") {
    return {
      title: "Slack account connected",
      description: "Your account is already connected with Slack.",
    };
  }
  switch (status) {
    case "success":
      return {
        title: "Slack account connected",
        description:
          "You can return to Slack and use FortyOne commands, message actions, and Maya.",
      };
    case "error":
      return {
        title: "Slack account not connected",
        description:
          errorMessage ??
          "FortyOne could not connect this account. Try the link again.",
      };
    case "linking":
      return {
        title: "Connecting your Slack account",
        description:
          "FortyOne is securely matching your Slack identity to this account.",
      };
    default:
      return {
        title: "Connect your Slack account",
        description: hasToken
          ? "Your secure Slack link is ready to be connected."
          : "Open the private connection link sent to you by FortyOne in Slack.",
      };
  }
};

const getConnectionMethod = (linkedVia: string) => {
  switch (linkedVia) {
    case "email_match":
      return "Automatic email match";
    case "dashboard_oauth":
      return "Slack authorization";
    case "manual_link":
      return "Slack connection link";
    default:
      return "Slack connection";
  }
};

export const SlackAccountLinkSettings = () => {
  const { withWorkspace } = useWorkspacePath();
  const { data: integration } = useSlackIntegration();
  const { errorMessage, hasToken, retry, status } = useSlackAccountLinkToken();
  const createAccountLinkSession = useCreateSlackAccountLinkSession();
  const disconnectAccount = useDisconnectSlackAccount();
  const [isDisconnectOpen, setIsDisconnectOpen] = useState(false);
  const accountLink = integration?.accountLink;
  const slackWorkspace = integration?.slackWorkspace;
  const hasConfirmedLink =
    Boolean(accountLink) ||
    status === "success" ||
    status === "already_connected";
  const { description, title } = getAccountLinkCopy({
    errorMessage,
    hasToken,
    isLinked: hasConfirmedLink,
    status,
  });

  const connectAccount = () => {
    createAccountLinkSession.mutate(window.location.href, {
      onSuccess: (response) => {
        if (response.error?.message) {
          toast.error("Slack", { description: response.error.message });
          return;
        }
        if (response.data?.linked) {
          toast.info("Your account is already connected with Slack");
          return;
        }
        if (response.data?.installUrl) {
          window.location.assign(response.data.installUrl);
          return;
        }
        toast.error("Slack", {
          description: "FortyOne could not start Slack account linking.",
        });
      },
    });
  };

  return (
    <Box>
      <Flex align="center" className="mb-6" gap={2}>
        <SettingsBackButton
          href={withWorkspace("/settings/integrations")}
          label="Back to integrations"
        />
        <Text as="h1" className="text-2xl font-medium">
          Slack
        </Text>
      </Flex>

      <Box className="border-border bg-surface rounded-2xl border p-6">
        <Flex align="center" gap={3}>
          <Flex
            align="center"
            className="bg-surface-muted size-10 shrink-0 rounded-lg"
            justify="center"
          >
            <SlackIcon className="h-5 w-5" />
          </Flex>
          <Box>
            <Text className="font-medium">{title}</Text>
            <Text className="mt-1" color="muted">
              {description}
            </Text>
          </Box>
        </Flex>

        {accountLink ? (
          <Box className="border-border bg-surface-muted mt-6 rounded-xl border px-4 py-3">
            <Flex align="center" justify="between">
              <Text color="muted">Slack workspace</Text>
              <Text className="font-medium">
                {slackWorkspace?.slackTeamName ?? "Connected workspace"}
              </Text>
            </Flex>
            <Flex align="center" className="mt-2" justify="between">
              <Text color="muted">Slack member ID</Text>
              <Text className="font-medium">{accountLink.slackUserId}</Text>
            </Flex>
            <Flex align="center" className="mt-2" justify="between">
              <Text color="muted">Connection method</Text>
              <Text className="font-medium">
                {getConnectionMethod(accountLink.linkedVia)}
              </Text>
            </Flex>
            <Flex align="center" className="mt-2" justify="between">
              <Text color="muted">Connected</Text>
              <Text className="font-medium">
                {format(new Date(accountLink.linkedAt), "MMM d, yyyy")}
              </Text>
            </Flex>
          </Box>
        ) : null}

        <Flex align="center" className="mt-6" gap={3}>
          {status === "error" && retry ? (
            <Button onClick={retry}>Try again</Button>
          ) : null}
          {!hasConfirmedLink && status !== "linking" ? (
            <Button
              color="invert"
              loading={createAccountLinkSession.isPending}
              onClick={connectAccount}
            >
              Connect Slack account
            </Button>
          ) : null}
          {accountLink ? (
            <Button
              color="tertiary"
              leftIcon={<UnlinkIcon />}
              onClick={() => {
                setIsDisconnectOpen(true);
              }}
            >
              Disconnect account
            </Button>
          ) : null}
        </Flex>
      </Box>

      <Dialog onOpenChange={setIsDisconnectOpen} open={isDisconnectOpen}>
        <Dialog.Content>
          <Dialog.Header>
            <Dialog.Title className="px-6 pt-0.5 text-lg">
              Disconnect Slack account
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Text color="muted">
              Maya, Slack commands, and Slack actions will stop recognizing you
              as this FortyOne account. You can reconnect later.
            </Text>
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-3 border-0 pt-2">
            <Button
              color="tertiary"
              onClick={() => {
                setIsDisconnectOpen(false);
              }}
            >
              Cancel
            </Button>
            <Button
              loading={disconnectAccount.isPending}
              onClick={() => {
                disconnectAccount.mutate(undefined, {
                  onSuccess: (response) => {
                    if (!response.error) setIsDisconnectOpen(false);
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
