"use client";

import Link from "next/link";
import { Box, Button, Flex, Text } from "ui";
import { SlackIcon } from "icons";
import type { SlackAccountLinkStatus } from "@/lib/hooks/slack/use-account-link-token";
import { useWorkspacePath } from "@/hooks";
import { useSlackAccountLinkToken } from "@/lib/hooks/slack";

const getAccountLinkCopy = ({
  errorMessage,
  hasToken,
  status,
}: {
  errorMessage: string | null;
  hasToken: boolean;
  status: SlackAccountLinkStatus;
}) => {
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

export const SlackAccountLinkSettings = () => {
  const { withWorkspace } = useWorkspacePath();
  const { errorMessage, hasToken, retry, status } = useSlackAccountLinkToken();
  const { description, title } = getAccountLinkCopy({
    errorMessage,
    hasToken,
    status,
  });

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Slack
      </Text>

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

        <Flex align="center" className="mt-6" gap={3}>
          {status === "error" && retry ? (
            <Button onClick={retry}>Try again</Button>
          ) : null}
          <Link href={withWorkspace("/settings/integrations")}>
            <Text color="muted">Back to integrations</Text>
          </Link>
        </Flex>
      </Box>
    </Box>
  );
};
