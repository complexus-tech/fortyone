import Link from "next/link";
import { Box, Button, Flex, Text } from "ui";
import type { User } from "@/types";
import { Logo } from "@/components/ui";
import { ProfileForm } from "@/modules/settings/account/profile/components/form";
import { PublicPortalShell } from "./portal-shell";
import type { PublicPortal, PublicPortalViewer } from "./types";
import { PublicPortalUserMenu } from "./user-menu";

const AccountContent = ({
  profile,
  showWorkspaceSetup = false,
}: {
  profile: User;
  showWorkspaceSetup?: boolean;
}) => (
  <Box className="mx-auto w-full max-w-3xl px-4 py-10 md:px-6 md:py-14">
    <Box className="mb-8">
      <Text as="h1" className="text-2xl" fontWeight="semibold">
        Account settings
      </Text>
      <Text className="mt-2 max-w-xl" color="muted">
        Manage the profile people see when you post feedback and join public
        discussions.
      </Text>
    </Box>
    <ProfileForm initialProfile={profile} />
    {showWorkspaceSetup ? (
      <Flex
        align="center"
        className="border-border bg-surface mt-6 flex-col items-start rounded-2xl border p-6 sm:flex-row sm:items-center"
        gap={6}
        justify="between"
      >
        <Box>
          <Text fontWeight="semibold">Bring your own team to FortyOne</Text>
          <Text className="mt-1" color="muted">
            Create a workspace when you are ready to manage projects with your
            team.
          </Text>
        </Box>
        <Button className="shrink-0" color="invert" href="/onboarding/create">
          Create workspace
        </Button>
      </Flex>
    ) : null}
  </Box>
);

export const PublicPortalAccountPage = ({
  portal,
  profile,
  viewer,
}: {
  portal: PublicPortal;
  profile: User;
  viewer: PublicPortalViewer;
}) => (
  <PublicPortalShell portal={portal} viewer={viewer}>
    <AccountContent profile={profile} />
  </PublicPortalShell>
);

export const AccountPage = ({
  profile,
  viewer,
}: {
  profile: User;
  viewer: PublicPortalViewer;
}) => (
  <Box className="bg-background min-h-dvh">
    <Box className="border-border/60 bg-background sticky top-0 z-20 border-b">
      <Flex
        align="center"
        className="mx-auto h-16 w-full max-w-5xl px-4 md:px-6"
        justify="between"
      >
        <Link
          aria-label="FortyOne home"
          className="transition-opacity hover:opacity-80"
          href={viewer.appHref ?? "/account"}
        >
          <Logo className="h-8" />
        </Link>
        <PublicPortalUserMenu viewer={viewer} />
      </Flex>
    </Box>
    <AccountContent profile={profile} showWorkspaceSetup={!viewer.appHref} />
  </Box>
);
