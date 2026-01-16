import { Box, Button, Text } from "ui";
import { Logo } from "@/components/ui";
import type { Invitation } from "@/modules/invitations/types";
import { auth } from "@/auth";
import { JoinForm } from "./components/join-form";

export const JoinWorkspace = async ({
  invitation,
}: {
  invitation: Invitation;
}) => {
  const session = await auth();
  const { email, workspaceName, role } = invitation;
  const canJoin = session?.user?.email === email;

  return (
    <Box className="w-full px-6 md:max-w-md md:px-0">
      <Logo asIcon />
      <Text as="h1" className="mt-10 mb-6 text-4xl" fontWeight="semibold">
        Your invitation to {workspaceName}
      </Text>
      <Text className="mb-6" color="muted">
        You&apos;ve been invited to join the team at{" "}
        <span className="font-semibold">{workspaceName}</span> as{" "}
        {role === "admin" ? "an" : "a"}{" "}
        <span className="font-semibold capitalize">{role}.</span>
        {!canJoin ? (
          <>
            {" "}
            Sign in with the email{" "}
            <span className="font-semibold">{email}</span> to continue.
          </>
        ) : null}
      </Text>
      {canJoin ? (
        <JoinForm invitation={invitation} />
      ) : (
        <Button
          align="center"
          className="md:py-3"
          color="invert"
          fullWidth
          href="/login"
          size="lg"
        >
          Sign in
        </Button>
      )}
    </Box>
  );
};
