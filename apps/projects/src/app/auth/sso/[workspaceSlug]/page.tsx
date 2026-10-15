import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { OnboardingLayout } from "@/components/layouts/onboarding-layout";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import { isDesktopAuthPath } from "@/lib/desktop-auth";
import { WorkspaceSSORecovery } from "@/modules/auth/workspace-sso-recovery";
import { getWorkspaceSSOStatus } from "@/modules/auth/workspace-sso-status";
import { getLoginUrl, withCallbackUrl } from "@/utils/callback-url";

export const metadata: Metadata = {
  title: "Workspace sign-in - FortyOne",
  robots: { index: false, follow: false },
};

export default async function WorkspaceSSOPage({
  params,
  searchParams,
}: {
  params: Promise<{ workspaceSlug: string }>;
  searchParams?: Promise<{ callbackUrl?: string }>;
}) {
  const [{ workspaceSlug }, session, query] = await Promise.all([
    params,
    auth(),
    searchParams,
  ]);
  const callbackUrl = isDesktopAuthPath(query?.callbackUrl)
    ? query?.callbackUrl
    : undefined;
  const path = withCallbackUrl(
    `/auth/sso/${encodeURIComponent(workspaceSlug)}`,
    callbackUrl,
  );
  if (!session) redirect(getLoginUrl(path));

  const [workspaces, status] = await Promise.all([
    getWorkspaces(),
    getWorkspaceSSOStatus(workspaceSlug),
  ]);
  const workspace = workspaces.find(
    (candidate) => candidate.slug.toLowerCase() === workspaceSlug.toLowerCase(),
  );
  if (!workspace) redirect("/unauthorized");
  if (!status.enabled)
    redirect(callbackUrl ?? `/${encodeURIComponent(workspace.slug)}`);

  return (
    <OnboardingLayout>
      <WorkspaceSSORecovery
        callbackUrl={callbackUrl}
        email={session.user.email}
        required={status.requireSSO}
        workspaceName={workspace.name}
        workspaceSlug={workspace.slug}
      />
    </OnboardingLayout>
  );
}
