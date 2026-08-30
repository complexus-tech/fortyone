"use client";

import { ChatIcon, GitHubIcon, SlackIcon } from "icons";
import type { IntegrationRequest } from "../types";
import type { RequestSourceBannerDetails } from "./request-integration-banner";

export const getRequestSourceBanner = ({
  issueNumber,
  provider,
  repositoryName,
  slackChannel,
  storyTerm,
}: {
  issueNumber: string;
  provider: IntegrationRequest["provider"];
  repositoryName: string | null;
  slackChannel: string | null;
  storyTerm: string;
}): RequestSourceBannerDetails => {
  switch (provider) {
    case "github":
      return {
        icon: <GitHubIcon className="text-primary h-5 shrink-0" />,
        openLabel: "Open on GitHub",
        primaryText: `Issue synced with GitHub ${issueNumber}`.trim(),
        secondaryText: repositoryName,
      };
    case "slack":
      return {
        icon: <SlackIcon className="h-5 shrink-0" />,
        openLabel: "Open on Slack",
        primaryText: "Request from Slack",
        secondaryText: slackChannel ? `#${slackChannel}` : null,
      };
    case "intercom":
      return {
        icon: <ChatIcon className="text-primary h-5 shrink-0" />,
        openLabel: "Open source",
        primaryText: `${storyTerm} from Intercom`,
        secondaryText: null,
      };
  }
};
