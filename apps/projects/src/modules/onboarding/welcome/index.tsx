"use client";
import { Badge, Box, Button, Container, Text } from "ui";
import { CommandIcon, DownloadIcon, SettingsIcon } from "icons";
import { Logo } from "@/components/ui";
import type { User, Workspace } from "@/types";
import { ActionCard } from "./components/action-card";
import { getWelcomeDestinations } from "./destinations";

export const Welcome = ({
  callbackUrl,
  workspaces,
  profile,
}: {
  callbackUrl?: string;
  workspaces: Workspace[];
  profile: User;
}) => {
  const { importUrl, redirectUrl } = getWelcomeDestinations(
    workspaces,
    profile.lastUsedWorkspaceId,
    callbackUrl,
  );

  return (
    <Container className="max-w-md md:max-w-xl">
      <Logo asIcon />
      <Text as="h1" className="mt-10 mb-6 text-4xl" fontWeight="semibold">
        Welcome to FortyOne👋
      </Text>
      <Text className="mb-6" color="muted">
        Your workspace is ready. Create your first story by pressing{" "}
        <Badge
          className="inline-flex rounded-md font-semibold"
          color="tertiary"
        >
          shift+n
        </Badge>{" "}
        or explore the features below.
      </Text>
      <Box className="grid gap-4">
        {importUrl ? (
          <ActionCard
            description="Bring work from Jira or upload an export. You'll review everything before anything is created."
            href={importUrl}
            icon={<DownloadIcon />}
            title="Import existing work"
          />
        ) : null}
        <ActionCard
          description="Connect your GitHub repositories and Slack channels for seamless integration."
          href={redirectUrl}
          icon={<SettingsIcon />}
          title="Set up integrations"
        />
        <ActionCard
          description={
            <>
              Boost your productivity with keyboard shortcuts (press{" "}
              <Badge
                className="text-muted dark:text-text-secondary inline-flex gap-0 rounded-md font-semibold"
                color="tertiary"
              >
                <CommandIcon className="h-3" strokeWidth={3} />
                +k
              </Badge>{" "}
              for help)
            </>
          }
          href={redirectUrl}
          icon={<CommandIcon />}
          title="Master shortcuts"
        />
      </Box>
      <Button
        align="center"
        className="mt-4 md:py-3"
        color="invert"
        fullWidth
        href={redirectUrl}
        size="lg"
      >
        Get Started
      </Button>
    </Container>
  );
};
