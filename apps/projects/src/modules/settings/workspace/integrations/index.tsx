"use client";

import Link from "next/link";
import { useState, type ReactNode } from "react";
import { Box, Flex, Input, Text } from "ui";
import { cn } from "lib";
import { SearchIcon, SlackIcon } from "icons";
import { useGitHubIntegration } from "@/lib/hooks/github";
import { useSlackIntegration } from "@/lib/hooks/slack";
import { useWorkspacePath } from "@/hooks";

const GitHubIcon = () => (
  <svg
    className="h-8 w-8"
    fill="currentColor"
    viewBox="0 0 24 24"
    xmlns="http://www.w3.org/2000/svg"
  >
    <path d="M12 0C5.374 0 0 5.373 0 12c0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0 1 12 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.3 24 12c0-6.627-5.373-12-12-12z" />
  </svg>
);

const FigmaIcon = () => (
  <svg
    className="h-8 w-8"
    fill="none"
    viewBox="0 0 38 57"
    xmlns="http://www.w3.org/2000/svg"
  >
    <path
      d="M19 28.5C19 23.3 23.3 19 28.5 19S38 23.3 38 28.5 33.7 38 28.5 38 19 33.7 19 28.5Z"
      fill="#1ABCFE"
    />
    <path
      d="M0 47.5C0 42.3 4.3 38 9.5 38H19v9.5C19 52.7 14.7 57 9.5 57S0 52.7 0 47.5Z"
      fill="#0ACF83"
    />
    <path
      d="M19 0v19h9.5C33.7 19 38 14.7 38 9.5S33.7 0 28.5 0H19Z"
      fill="#FF7262"
    />
    <path
      d="M0 9.5C0 14.7 4.3 19 9.5 19H19V0H9.5C4.3 0 0 4.3 0 9.5Z"
      fill="#F24E1E"
    />
    <path
      d="M0 28.5C0 33.7 4.3 38 9.5 38H19V19H9.5C4.3 19 0 23.3 0 28.5Z"
      fill="#A259FF"
    />
  </svg>
);

const GitLabIcon = () => (
  <svg
    className="h-8 w-8"
    viewBox="0 0 380 380"
    xmlns="http://www.w3.org/2000/svg"
  >
    <path d="M190 350L250 170H130L190 350Z" fill="#E24329" />
    <path d="M190 350L130 170H30L190 350Z" fill="#FC6D26" />
    <path
      d="M30 170L10 230C8 237 11 245 17 249L190 350L30 170Z"
      fill="#FCA326"
    />
    <path d="M30 170H130L90 30C88 23 78 23 76 30L30 170Z" fill="#E24329" />
    <path d="M190 350L250 170H350L190 350Z" fill="#FC6D26" />
    <path
      d="M350 170L370 230C372 237 369 245 363 249L190 350L350 170Z"
      fill="#FCA326"
    />
    <path
      d="M350 170H250L290 30C292 23 302 23 304 30L350 170Z"
      fill="#E24329"
    />
  </svg>
);

const IntercomIcon = () => (
  <svg
    className="h-8 w-8"
    viewBox="0 0 28 28"
    xmlns="http://www.w3.org/2000/svg"
  >
    <rect fill="#1F8DED" height="28" rx="5" width="28" />
    <path
      d="M22.17 18.86c-.1 0-5.57 3.84-8.17 3.84s-8.07-3.84-8.17-3.84a.78.78 0 0 0-1.05.28.72.72 0 0 0 .29 1.01c.42.27 5.2 4.31 8.93 4.31s8.51-4.04 8.93-4.31a.72.72 0 0 0 .29-1.01.78.78 0 0 0-1.05-.28zM7.24 16.06a.72.72 0 0 0 .72-.72V7.87a.72.72 0 0 0-1.44 0v7.47c0 .4.32.72.72.72zm3.59 1.57a.72.72 0 0 0 .72-.72V6.04a.72.72 0 0 0-1.44 0v10.87c0 .4.32.72.72.72zm3.59 0a.72.72 0 0 0 .72-.72V6.04a.72.72 0 0 0-1.44 0v10.87c0 .4.32.72.72.72zm3.59-1.57a.72.72 0 0 0 .72-.72V7.87a.72.72 0 0 0-1.44 0v7.47c0 .4.32.72.72.72zm3.6-1.2a.72.72 0 0 0 .71-.72V9.07a.72.72 0 0 0-1.44 0v5.07c0 .4.33.72.72.72z"
      fill="#fff"
    />
  </svg>
);

type Integration = {
  id: string;
  name: string;
  description: string;
  icon: ReactNode;
  enabled: boolean;
  href?: string;
};

const getIntegrationStatus = (integration: Integration) => {
  if (!integration.href) return "Coming soon";
  return integration.enabled ? "Installed" : "Not connected";
};

const IntegrationCard = ({
  integration,
  basePath,
  showDescription = true,
}: {
  integration: Integration;
  basePath: string;
  showDescription?: boolean;
}) => {
  const content = (
    <Box
      className={cn(
        "border-border rounded-xl border p-5 transition-colors",
        integration.href
          ? "hover:border-text-muted/30"
          : "cursor-not-allowed opacity-50",
      )}
    >
      <Flex align="center" gap={3}>
        {integration.icon}
        <Box>
          <Text className="font-medium">{integration.name}</Text>
          <Text color="muted">{getIntegrationStatus(integration)}</Text>
        </Box>
      </Flex>
      {showDescription ? (
        <Text className="mt-3 line-clamp-2" color="muted">
          {integration.description}
        </Text>
      ) : null}
    </Box>
  );

  if (integration.href) {
    return <Link href={`${basePath}/${integration.href}`}>{content}</Link>;
  }

  return content;
};

export const IntegrationsIndex = () => {
  const { data: integration } = useGitHubIntegration();
  const { data: slackIntegration } = useSlackIntegration();
  const { withWorkspace } = useWorkspacePath();
  const [search, setSearch] = useState("");

  const basePath = withWorkspace("/settings/workspace/integrations");
  const isGitHubEnabled = (integration?.installations.length ?? 0) > 0;
  const isSlackEnabled = Boolean(slackIntegration?.slackWorkspace?.isActive);

  const allIntegrations: Integration[] = [
    {
      id: "github",
      name: "GitHub",
      description: "Link PRs, branches, and commits to stories.",
      icon: <GitHubIcon />,
      enabled: isGitHubEnabled,
      href: "github",
    },
    {
      id: "slack",
      name: "Slack",
      description: "Turn conversations into stories from Slack.",
      icon: <SlackIcon className="h-8 w-8" />,
      enabled: isSlackEnabled,
      href: "slack",
    },
    {
      id: "gitlab",
      name: "GitLab",
      description: "Connect merge requests and pipelines to stories.",
      icon: <GitLabIcon />,
      enabled: false,
    },
    {
      id: "figma",
      name: "Figma",
      description: "Attach designs to stories from Figma.",
      icon: <FigmaIcon />,
      enabled: false,
    },
    {
      id: "intercom",
      name: "Intercom",
      description: "Route customer conversations into stories.",
      icon: <IntercomIcon />,
      enabled: false,
    },
  ];

  const enabledIntegrations = allIntegrations.filter((i) => i.enabled);
  const query = search.trim().toLowerCase();
  const filteredIntegrations = query
    ? allIntegrations.filter((i) => i.name.toLowerCase().includes(query))
    : allIntegrations;
  const filteredEnabled = query
    ? enabledIntegrations.filter((i) => i.name.toLowerCase().includes(query))
    : enabledIntegrations;

  return (
    <Box>
      <Text as="h1" className="mb-2 text-2xl font-medium">
        Integrations
      </Text>
      <Text color="muted">
        Enhance your workspace with a wide variety of add-ons and integrations.
      </Text>

      <Box className="mt-6">
        <Input
          leftIcon={<SearchIcon className="h-4.5" />}
          onChange={(e) => {
            setSearch(e.target.value);
          }}
          placeholder="Search integrations"
          rounded="xl"
          size="lg"
          value={search}
        />
      </Box>

      {filteredEnabled.length > 0 && (
        <Box className="mt-8">
          <Text
            className="mb-4 text-sm font-medium tracking-wider uppercase"
            color="muted"
          >
            Installed
          </Text>
          <Box className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {filteredEnabled.map((integration) => (
              <IntegrationCard
                basePath={basePath}
                integration={integration}
                key={integration.id}
                showDescription={false}
              />
            ))}
          </Box>
        </Box>
      )}

      <Box className="mt-8">
        <Text
          className="mb-4 text-sm font-medium tracking-wider uppercase"
          color="muted"
        >
          Essentials
        </Text>
        <Box className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {filteredIntegrations.map((integration) => (
            <IntegrationCard
              basePath={basePath}
              integration={integration}
              key={integration.id}
            />
          ))}
        </Box>
      </Box>
    </Box>
  );
};
