import type { Metadata } from "next";
import { Box, Button, Text } from "ui";
import { JoinWorkspace } from "@/modules/onboarding/join";
import { verifyInvitation } from "@/lib/queries/verify-invitation";
import { Logo } from "@/components/ui";

export const metadata: Metadata = {
  title: "Join Workspace - FortyOne",
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
      <Box className="w-full px-6 md:max-w-md md:px-0">
        <Logo asIcon />
        <Text
          as="h1"
          className="mt-10 mb-6 text-4xl first-letter:uppercase"
          fontWeight="semibold"
        >
          {res.error.message}
        </Text>
        <Text className="mb-6" color="muted">
          {res.error.message.includes("expired")
            ? "The invitation has expired. Please contact your team to get a new invitation. New invitations are valid for 24 hours."
            : "If you believe this is an error, please contact your team to get a new invitation."}
        </Text>
        <Button
          align="center"
          className="md:py-3"
          color="invert"
          fullWidth
          href="/"
          size="lg"
        >
          Go back
        </Button>
      </Box>
    );
  }

  return <JoinWorkspace invitation={res.data!} />;
}
