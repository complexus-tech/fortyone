"use client";
import { Badge, Box, Button, Container, Text } from "ui";
import { CommandIcon, SettingsIcon, TeamIcon } from "icons";
import { Logo } from "@/components/ui";
import type { User, Workspace } from "@/types";
import { ActionCard } from "./components/action-card";

const domain = process.env.NEXT_PUBLIC_DOMAIN!;

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
  return `https://${activeWorkspace.slug}.${domain}/my-work`;
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
    <Container className="max-w-md md:max-w-lg">
      <Logo />
      <Text as="h1" className="mb-2 mt-6 text-[1.7rem]" fontWeight="semibold">
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
                className="inline-flex gap-0 rounded-[0.4rem] font-semibold text-gray dark:text-gray-300"
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
        href={redirectUrl}
      >
        Get Started
      </Button>
    </Container>
  );
};
