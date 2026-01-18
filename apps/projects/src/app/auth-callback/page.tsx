import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { getMyInvitations } from "@/lib/queries/get-invitations";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getProfile } from "@/lib/queries/profile";
import { ClientPage } from "./client";
import { getAuthCode } from "@/lib/queries/get-auth-code";

export default async function AuthCallback({
  searchParams,
}: {
  searchParams: Promise<{ mobileApp?: string }>;
}) {
  const params = await searchParams;
  const isMobileApp = params?.mobileApp === "true";

  const session = await auth();

  if (!session) {
    redirect("/");
  }

  const [invitations, workspaces, profile] = await Promise.all([
    getMyInvitations(),
    getWorkspaces(session.token),
    getProfile(session),
  ]);

  if (isMobileApp && workspaces.length > 0) {
    const authCodeResponse = await getAuthCode(session);
    if (authCodeResponse.error || !authCodeResponse.data) {
      redirect("/?mobileApp=true&error=Failed to generate auth code");
    } else {
      redirect(
        `fortyone://login?code=${authCodeResponse.data.code}&email=${authCodeResponse.data.email}`,
      );
    }
  }
  return (
    <ClientPage
      invitations={invitations.data || []}
      profile={profile}
      session={session}
      workspaces={workspaces}
    />
  );
}
