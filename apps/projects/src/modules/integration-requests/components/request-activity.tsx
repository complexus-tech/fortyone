"use client";

import { ClockIcon, GitHubIcon, SlackIcon } from "icons";
import { Box, Tabs, Text } from "ui";
import type { IntegrationRequestProvider } from "../types";
import { IntegrationRequestThreadActivity } from "../thread-activity";
import { RequestGitHubComments } from "./request-github-comments";

export const RequestActivity = ({
  provider,
  requestId,
}: {
  provider: IntegrationRequestProvider;
  requestId: string;
}) => {
  if (provider === "github") {
    return (
      <Box>
        <Text
          as="h4"
          className="mb-4 flex items-center gap-1"
          fontWeight="medium"
        >
          <ClockIcon className="relative -top-px" />
          Activity feed
        </Text>
        <Tabs defaultValue="github">
          <Tabs.List className="mx-0 mb-5 md:mx-0">
            <Tabs.Tab
              className="gap-1 px-2"
              leftIcon={<GitHubIcon className="h-[1.05rem]" />}
              value="github"
            >
              GitHub
            </Tabs.Tab>
          </Tabs.List>
          <Tabs.Panel value="github">
            <RequestGitHubComments requestId={requestId} />
          </Tabs.Panel>
        </Tabs>
      </Box>
    );
  }

  if (provider === "slack") {
    return (
      <Box>
        <Text
          as="h4"
          className="mb-4 flex items-center gap-1"
          fontWeight="medium"
        >
          <ClockIcon className="relative -top-px" />
          Activity feed
        </Text>
        <Tabs defaultValue="slack">
          <Tabs.List className="mx-0 mb-5">
            <Tabs.Tab
              className="gap-1 px-2"
              leftIcon={<SlackIcon className="h-[1.05rem]" />}
              value="slack"
            >
              Slack
            </Tabs.Tab>
          </Tabs.List>
          <Tabs.Panel value="slack">
            <IntegrationRequestThreadActivity requestId={requestId} />
          </Tabs.Panel>
        </Tabs>
      </Box>
    );
  }

  return <Text color="muted">No integration activity is available.</Text>;
};
