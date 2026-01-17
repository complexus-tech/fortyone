import type { ReactNode } from "react";
import { ArrowLeft2Icon, TeamIcon } from "icons";
import { Box, Button, Text } from "ui";
import { auth } from "@/auth";
import { getTeam } from "@/modules/teams/queries/get-team";

export default async function Layout({
  children,
  params,
}: {
  children: ReactNode;
  params: Promise<{ teamId: string; workspaceSlug: string }>;
}) {
  const { teamId, workspaceSlug } = await params;
  const session = await auth();
  const ctx = { session: session!, workspaceSlug };
  const data = await getTeam(teamId, ctx);
  if (data.error?.message) {
    return (
      <Box className="flex h-screen items-center justify-center">
        <Box className="flex flex-col items-center">
          <TeamIcon className="h-16 w-auto" />
          <Text className="mb-6 mt-10" fontSize="3xl">
            404: Team not found
          </Text>
          <Text className="mb-6 max-w-md text-center" color="muted">
            This team might not exist or you do not belong to this team.
          </Text>
          <Button
            className="gap-1 pl-2"
            color="tertiary"
            href="/my-work"
            leftIcon={<ArrowLeft2Icon className="h-[1.05rem] w-auto" />}
          >
            Go to my work
          </Button>
        </Box>
      </Box>
    );
  }
  return <>{children}</>;
}
