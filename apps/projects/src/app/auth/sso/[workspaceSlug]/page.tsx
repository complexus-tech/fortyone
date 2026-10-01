import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { OnboardingLayout } from "@/components/layouts/onboarding-layout";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import { WorkspaceSSORecovery } from "@/modules/auth/workspace-sso-recovery";
import { getWorkspaceSSOStatus } from "@/modules/auth/workspace-sso-status";
import { getLoginUrl } from "@/utils/callback-url";

export const metadata: Metadata = {
  title: "Workspace sign-in - FortyOne",
  robots: { index: false, follow: false },
};

export default async function WorkspaceSSOPage({
  params,
}: {
  params: Promise<{ workspaceSlug: string }>;
}) {
  const [{ workspaceSlug }, session] = await Promise.all([params, auth()]);
  const path = `/auth/sso/${encodeURIComponent(workspaceSlug)}`;
  if (!session) redirect(getLoginUrl(path));

  const [workspaces, status] = await Promise.all([
    getWorkspaces(),
    getWorkspaceSSOStatus(workspaceSlug),
  ]);
  const workspace = workspaces.find(
    (candidate) => candidate.slug.toLowerCase() === workspaceSlug.toLowerCase(),
  );
  if (!workspace) redirect("/unauthorized");
  if (!status.enabled) redirect(`/${encodeURIComponent(workspace.slug)}`);

  return (
    <OnboardingLayout>
      <WorkspaceSSORecovery
        email={session.user.email}
        required={status.requireSSO}
        workspaceName={workspace.name}
        workspaceSlug={workspace.slug}
      />
    </OnboardingLayout>
  );
}
