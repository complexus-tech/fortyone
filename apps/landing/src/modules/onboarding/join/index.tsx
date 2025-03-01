import { Box, Button, Text } from "ui";
import { Logo } from "@/components/ui";
import type { Invitation } from "@/types";
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
    <Box className="w-full max-w-md">
      <Logo asIcon className="relative -left-2 h-10 text-white" />
      <Text as="h1" className="mb-2 mt-6 text-[1.7rem]" fontWeight="semibold">
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
        <Button align="center" className="md:py-3" fullWidth href="/login">
          Sign in
        </Button>
      )}
    </Box>
  );
};
