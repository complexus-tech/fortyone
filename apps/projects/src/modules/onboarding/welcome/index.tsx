"use client";
import { Badge, Box, Button, Container, Text } from "ui";
import { CommandIcon, SettingsIcon, TeamIcon } from "icons";
import { Logo } from "@/components/ui";
import type { User, Workspace } from "@/types";
import { ActionCard } from "./components/action-card";

const getRedirectUrl = (
  workspaces: Workspace[],
  lastUsedWorkspaceId?: string,
) => {
  const activeWorkspace =
    workspaces.find((workspace) => workspace.id === lastUsedWorkspaceId) ||
    workspaces[0];
  if (!activeWorkspace) {
    return "/onboarding/create";
  }
  return `/${activeWorkspace.slug}/maya`;
};

export const Welcome = ({
  workspaces,
  profile,
}: {
  workspaces: Workspace[];
  profile: User;
}) => {
  const redirectUrl = getRedirectUrl(workspaces, profile?.lastUsedWorkspaceId);

  return (
    <Container className="max-w-md md:max-w-xl">
      <Logo asIcon />
      <Text as="h1" className="mt-10 mb-6 text-4xl" fontWeight="semibold">
        Welcome to FortyOne👋
      </Text>
      <Text className="mb-6" color="muted">
        Your workspace is ready. Create your first story by pressing{" "}
        <Badge
          className="inline-flex rounded-[0.4rem] font-semibold"
          color="tertiary"
        >
          shift+n
        </Badge>{" "}
        or explore the features below.
      </Text>
      <Box className="grid gap-4">
        <ActionCard
          description="Collaborate better by inviting your teammates to join your workspace."
          href={redirectUrl}
          icon={<TeamIcon />}
          title="Build your team"
        />
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
                className="text-muted dark:text-text-secondary inline-flex gap-0 rounded-[0.4rem] font-semibold"
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
        className="mt-4"
        color="invert"
        fullWidth
        size="lg"
        href={redirectUrl}
      >
        Get Started
      </Button>
    </Container>
  );
};
