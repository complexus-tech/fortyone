import type { Metadata } from "next";
import { Box, Button, Text } from "ui";
import { JoinWorkspace } from "@/modules/onboarding/join";
import { verifyInvitation } from "@/lib/actions/verify-invitation";
import { Logo } from "@/components/ui";

export const metadata: Metadata = {
  title: "Join Workspace - Complexus",
  description: "Join a workspace",
};

export default async function JoinWorkspacePage({
  searchParams,
}: {
  searchParams: Promise<{ token: string }>;
}) {
  const { token } = await searchParams;
  const res = await verifyInvitation(token);

  if (res.error?.message) {
    return (
      <Box className="w-full max-w-md">
        <Logo asIcon className="relative -left-2 h-10 text-white" />
        <Text as="h1" className="mb-2 mt-6 text-[1.7rem]" fontWeight="semibold">
          Invalid invitation
          {res.error.message}
        </Text>
        <Text className="mb-6" color="muted">
          The invitation is invalid or has expired. Please contact your team to
          get a new invitation.
        </Text>
        <Button href="/">Go to home</Button>
      </Box>
    );
  }

  return <JoinWorkspace invitation={res.data!} />;
}
