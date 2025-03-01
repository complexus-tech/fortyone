import { Badge, Box, Button, Text } from "ui";
import { CommandIcon, SettingsIcon, TeamIcon } from "icons";
import type { Session } from "next-auth";
import { Logo } from "@/components/ui";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { auth } from "@/auth";
import { ActionCard } from "./components/action-card";

const domain = process.env.NEXT_PUBLIC_DOMAIN!;

const getRedirectUrl = async (session: Session) => {
  const workspaces = (await getWorkspaces(session.token)).sort(
    (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime(),
  );
  const activeWorkspace = workspaces[0] || session.activeWorkspace;
  if (domain.includes("localhost")) {
    return `http://${activeWorkspace.slug}.localhost:3000/my-work`;
  }
  return `https://${activeWorkspace.slug}.${domain}/my-work`;
};

export const Welcome = async () => {
  const session = await auth();
  const redirectUrl = await getRedirectUrl(session!);

  return (
    <Box className="max-w-sm">
      <Logo asIcon className="relative -left-2 h-10 text-white" />
      <Text as="h1" className="mb-2 mt-6 text-[1.7rem]" fontWeight="semibold">
        Welcome to Complexus👋
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
          icon={<TeamIcon />}
          title="Build your team"
        />
        <ActionCard
          description="Connect your GitHub repositories and Slack channels for seamless integration."
          icon={<SettingsIcon />}
          title="Set up integrations"
        />
        <ActionCard
          description={
            <>
              Boost your productivity with keyboard shortcuts (press{" "}
              <Badge
                className="inline-flex rounded-[0.4rem] font-semibold"
                color="tertiary"
              >
                ?
              </Badge>{" "}
              for help)
            </>
          }
          icon={<CommandIcon />}
          title="Master shortcuts"
        />
      </Box>
      <Button align="center" className="mt-4" fullWidth href={redirectUrl}>
        Get Started
      </Button>
    </Box>
  );
};
