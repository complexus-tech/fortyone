import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { getMyInvitations } from "@/lib/queries/get-invitations";
import { ClientPage } from "./client";

export default async function AuthCallback() {
  const session = await auth();
  if (!session) {
    redirect("/login");
  }
  const invitations = await getMyInvitations();
  return <ClientPage invitations={invitations.data || []} session={session} />;
}
