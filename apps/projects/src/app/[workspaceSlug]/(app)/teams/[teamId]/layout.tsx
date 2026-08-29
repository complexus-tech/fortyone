import type { ReactNode } from "react";
import { auth } from "@/auth";
import { ResourceNotFoundState } from "@/components/ui/resource-not-found-state";
import { TeamEmptyIllustration } from "@/components/ui/illustrations/empty-state-illustrations";
import { getTeam } from "@/modules/teams/queries/get-team";
import { withWorkspacePath } from "@/utils";

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
      <ResourceNotFoundState
        description="This team might not exist or you might not belong to it."
        href={withWorkspacePath("/my-work", workspaceSlug)}
        illustration={<TeamEmptyIllustration />}
        title="404: Team not found"
      />
    );
  }

  return <>{children}</>;
}
