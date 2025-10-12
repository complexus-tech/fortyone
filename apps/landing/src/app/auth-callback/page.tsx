import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { getMyInvitations } from "@/lib/queries/get-invitations";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getProfile } from "@/lib/queries/profile";
import { ClientPage } from "./client";

export default async function AuthCallback({
  searchParams,
}: {
  searchParams: Promise<{ mobile?: string }>;
}) {
  const params = await searchParams;
  const isMobile = params?.mobile === "true";
  const session = await auth();
  if (!session) {
    redirect("/login");
  }
  if (isMobile) {
    redirect("fortyone://login?code=test");
  }
  const [invitations, workspaces, profile] = await Promise.all([
    getMyInvitations(),
    getWorkspaces(session.token),
    getProfile(session),
  ]);
  return (
    <ClientPage
      invitations={invitations.data || []}
      profile={profile}
      session={session}
      workspaces={workspaces}
    />
  );
}
